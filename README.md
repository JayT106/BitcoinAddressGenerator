# bitcoinAddressGenerator
TBC...

### Requirements
[go](https://golang.org/) 1.15 or newer.

### Build from source
- Install pre-required package
```bash
sudo apt-get update
sudo apt-get install -y git cmake
```
- Clone project
```bash
git clone https://github.com/JayT106/bitcoinAddressGenerator.git
```
- Install Go according to the installation instructions here:
  http://golang.org/doc/install

- Ensure Go was installed properly and is a supported version:

```bash
go version
go env GOROOT GOPATH
```
NOTE: The `GOROOT` and `GOPATH` above must not be the same path.  It is
recommended that `GOPATH` is set to a directory in your home directory such as
`~/goprojects` to avoid write permission issues.  It is also recommended to add
`$GOPATH/bin` to your `PATH` at this point.

- Build the project
```bash
make
```
or 
```bash
make build
```
- Run the test suites
```bash
make tests
```
### Starting binary and running the examples
- After `make` built the project without any error, you can find the binary in the `bin` folder. Launch the server and the server will use `8080` as the HTTP listening port
```bash
cd bin
./bitcoinAddressGeneratorServer-1.0.0_linux_amd64
```
- Execute the scripts and binery in the example to know how to interactive with the server
- The seed file is in the `test` folder, it is a json format file. It can be the load when running:
```bash
cd example
./getServerPublicKey.sh
./genPublicKeyAndSegWitAddress [the output of the previous script] ../test/test.json 
```

## License
bitcoinAddressGenerator is MIT License
