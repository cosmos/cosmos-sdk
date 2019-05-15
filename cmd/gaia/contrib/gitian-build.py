#!/usr/bin/python3

import os
import sys
import argparse
import subprocess


GITIAN_REPO = 'https://github.com/devrandom/gitian-builder.git'
SIGS_REPO = 'https://github.com/cosmos/gaia.sigs.git'
CACHE_DIRNAME = '.gitian-builder-cache'
GO_DEBIAN_RELEASE = '1.12.5-1'
GO_TARBALL = "golang-debian-%s.tar.gz" % GO_DEBIAN_RELEASE
GO_TAR_REMOTE = "https://salsa.debian.org/go-team/compiler/golang/-/archive/debian/%s/%s" % (
    GO_DEBIAN_RELEASE,
    GO_TARBALL,
)
SIGN_COMMAND = 'gpg --detach-sign'


def check_deps():
    subprocess.check_call(['git', 'version'], stdout=subprocess.DEVNULL, stderr=subprocess.DEVNULL)
    subprocess.check_call(['ruby', '--version'], stdout=subprocess.DEVNULL, stderr=subprocess.DEVNULL)
    subprocess.check_call(['curl', '--version'], stdout=subprocess.DEVNULL, stderr=subprocess.DEVNULL)
    subprocess.check_call(['docker', 'version'], stdout=subprocess.DEVNULL, stderr=subprocess.DEVNULL)


def setup(workdir, cachedir, args, platforms):
    if not os.path.isdir('gaia.sigs'):
        subprocess.check_call(['git', 'clone', SIGS_REPO, 'gaia.sigs'])

    # clone and cache gitian-builder
    cached_gitian = os.path.join(cachedir, 'gitian-builder')
    if not os.path.isdir(cached_gitian):
        os.chdir(cachedir)
        subprocess.check_call(['git', 'clone', GITIAN_REPO, 'gitian-builder'])
        os.chdir('gitian-builder')
        make_image_prog = ['bin/make-base-vm', '--docker', '--suite', 'bionic', '--arch', 'amd64']
        subprocess.check_call(make_image_prog)
        os.chdir(workdir)

    # download and cache go sources tarball
    cached_go_tar = os.path.join(cachedir, GO_TARBALL)
    if not os.path.exists(cached_go_tar):
        subprocess.check_call(['curl', '-L', GO_TAR_REMOTE, '--output', cached_go_tar])

    # setup platform-specific build directories
    for platform in platforms:
        platformdir = 'gitian-builder-%s' % platform
        if not os.path.isdir(platformdir):
            subprocess.check_call(['cp', '-ar', cached_gitian, platformdir])
            inputsdir = os.path.join(platformdir, 'inputs')
            os.mkdir(inputsdir)
            subprocess.check_call(['cp', cached_go_tar, inputsdir])

    os.chdir(workdir)


def build(workdir, args, platforms, release, commit):
    for platform in platforms:
        descriptor = os.path.join(workdir, 'cmd', 'gaia', 'contrib', 'gitian-descriptors', 'gitian-%s.yml' % platform)
        commit = "cosmos-sdk=%s" % commit
        platformdir = 'gitian-builder-%s' % platform
        os.chdir(platformdir)
        subprocess.check_call(['bin/gbuild', '--commit', commit, descriptor])
        try:
            subprocess.check_call(['libexec/stop-target'])
        except subprocess.CalledProcessError as e:
            print("couldn't stop the target container: %s" % e, file=sys.stderr)
        if args.sign is not None:
            sign_verify(workdir, args, platform, release, descriptor)
        os.chdir(workdir)


def sign_verify(workdir, args, platform, release, descriptor):
    sigsdir = os.path.join(workdir, 'gaia.sigs')
    subprocess.check_call([
        'bin/gsign',
        '-p', SIGN_COMMAND,
        '-s', args.sign,
        '--destination=%s' % sigsdir,
        '--release=%s' % release,
        descriptor,
    ])
    subprocess.check_call([
        'bin/gverify',
        '--destination=%s' % sigsdir,
        '--release=%s' % release,
        descriptor,
    ])
    if args.commit_sigs:
        os.chdir(sigsdir)
        try:
            subprocess.check_call(['git', 'add', '.'])
            subprocess.check_call(['git', 'commit', '-m', 'Add %s reproducible build' % release])
        except subprocess.CalledProcessError as e:
            print("couldn't commit the signatures: %s" % e, file=sys.stderr)


if __name__ == '__main__':

    check_deps()

    parser = argparse.ArgumentParser(usage='%(prog)s [options]')
    parser.add_argument('-o', '--os', dest='os', default='dlw', help='specify which operating systems the '+
        'build is for. Default is %(default)s. l for Linux, w for Windows, d for Darwin')
    parser.add_argument('-s', '--sign', metavar='IDENTITY', dest='sign', help='sign binaries with IDENTITY')
    parser.add_argument('-S', '--commit-signatures', action='store_true', dest='commit_sigs', help='commit the signed builds to the gaia signatures repository')

    args = parser.parse_args()
    workdir = os.getcwd()
    cachedir = os.path.join(workdir, CACHE_DIRNAME)
    if not os.path.exists(cachedir):
        os.mkdir(cachedir)

    platforms = []
    if 'd' in args.os:
        platforms.append('darwin')
    if 'l' in args.os:
        platforms.append('linux')
    if 'l' in args.os:
        platforms.append('windows')

    setup(workdir, cachedir, args, platforms)

    os.environ['USE_DOCKER'] = '1'

    commit = subprocess.check_output(
        ['git', 'rev-parse', 'HEAD'],
        universal_newlines=True,
    ).strip()
    release = subprocess.check_output(
        ['git', 'describe', '--tags', '--abbrev=9'],
        universal_newlines=True,
    ).strip()

    build(workdir, args, platforms, release, commit)
