# this script is for generating protobuf files for the new google.golang.org/protobuf API
set -eo pipefail

echo "Cleaning API directory"
(cd api; find ./ -type f \( -iname \*.pulsar.go -o -iname \*.pb.go -o -iname \*.cosmos_orm.go -o -iname \*.pb.gw.go \) -delete; find . -empty -type d -delete; cd ..)

echo "Generating API module"
(cd proto; buf generate --template buf.gen.pulsar.yaml)

echo "Generate Pulsar Test Data"
(cd testutil/testdata; buf generate --template buf.gen.pulsar.yaml)

echo "Generate x/tx"
(cd x/tx; make codegen)