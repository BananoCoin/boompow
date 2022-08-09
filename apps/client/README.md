<p align="center">
  <img src="https://raw.githubusercontent.com/BananoCoin/boompow-next/master/logo.svg" width="300">
</p>

# Client

This is the BoomPoW client, it receives and calculates work requests from the BoomPoW server

## Usage

The BoomPoW binaries generate work using the GPU and CPU by default, if the GPU is not available, then it will use CPU only.

You can build a version with CPU only by following the compilation instructions below.

For AMD GPUs on linux, you will need to either use the `amdgpu-pro` driver or run the `amdgpu-installer` with the following:

```
amdgpu-install --usecase=opencl --no-dkms
```

## Compiling

### Windows

For windows, requirements are:

- [GOLang](https://go.dev/doc/install)
- [TDM-GCC](http://tdm-gcc.tdragon.net/download)
- [OCL_SDK_Light](https://github.com/GPUOpen-LibrariesAndSDKs/OCL-SDK/releases) (For OpenCL)

You can copy the `opencl.lib` to the `lib\x86_64` directory of TDM-GCC to compile with OpenCL support

To build:

```
go build -o boompow-client.exe -tags cl -ldflags "-X main.WSUrl=wss://boompow.banano.cc/ws/worker -X main.GraphQLURL=https://boompow.banano.cc/graphql" .
```

To build for CPU only remove `-tags cl`

### Linux

Each distribution will vary on requirements, but on a ubuntu/debian based build something like

```
sudo apt install golang-go build-essential ocl-icd-opencl-dev
```

should get you what you need.

Then you can run `./build.sh` to build the binary

### MacOS

MacOS is the same process as Linux, except you do not need to install OpenCL headers.
