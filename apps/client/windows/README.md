# BoomPoW Client

To use the client, simply run `boompow-client.exe`

To edit default options, edit the `run-with-options.bat` script with your desired options and run that instead.

## Registering to provide work

If you want to provide work (and get paid for it), register by running `register-provider.bat`

## Registering to request work

If you want to use BoomPoW to provide work for your service, register by running `register-service.bat`

Your account will be manually reviewed, after approved run `get-service-token.bat`

You will receive a token like `service:blahblah` - this token can be used to invoke the `workGenerate` mutation on the server.
