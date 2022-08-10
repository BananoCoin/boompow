# BoomPoW Client

To use the client, simply run it with

```
./boompow-client
```

To see options:

```
./boompow-client -help
```

## Registering to provide work

If you want to provide work (and get paid for it), register with

```
./boompow-client -register-provider
```

## Registering to request work

If you want to use BoomPoW to provide work for your service, start by creating an account

```
./boompow-client -register-service
```

Your account will be manually reviewed, after approved run

```
./boompow-client -generate-service-token`
```

You will receive a token like `service:blahblah` - this token can be used to invoke the `workGenerate` mutation on the server.
