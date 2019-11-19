# Stackdriver WebHook Handler in Cloud Run

Simple Cloud Run service to process Stackdriver notifications. You can use it to create channel (WebHook) and define policies (see `policy/` for examples) to create threshold policies to invoke channel when reached.

## Prerequisites

If you don't have one already, start by creating new project and configuring [Google Cloud SDK](https://cloud.google.com/sdk/docs/). Similarly, if you have not done so already, you will have [set up Cloud Run](https://cloud.google.com/run/docs/setup).


## Configuration

[bin/config](bin/config)

```
bin/config
```

## Stackdriver

### Channel

[bin/channel](bin/channel)

```
bin/channel
```

### Policy

[bin/policy](bin/policy)

```
bin/policy
```

## Cloud Run

[bin/deploy](bin/deploy)

```
bin/deploy
```

## Building Image

```shell
bin/image
```

## Cleanup

To cleanup all resources created by this sample execute

```shell
bin/cleanup
```

## Disclaimer

This is my personal project and it does not represent my employer. I take no responsibility for issues caused by this code. I do my best to ensure that everything works, but if something goes wrong, my apologies is all you will get.


