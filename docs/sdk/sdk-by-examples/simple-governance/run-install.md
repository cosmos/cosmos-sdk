## Installation

Once you have finallized your application, install it using `go get`. The following commands will install the pre-built modules and examples of the SDK as well as your `simpleGov` application:

```bash
go get github.com/<your_username>/cosmos-sdk
cd $GOPATH/src/github.com/<your_username>/cosmos-sdk
make get_vendor_deps
make install
make install_examples
```

Check that the app is correctly installed by typing:

```bash
simplegovcli -h
simplegovd -h
```