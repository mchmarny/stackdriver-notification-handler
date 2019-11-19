# Generic Stackdriver WebHook Handler implemented in Cloud Run

Simple Cloud Run service to handle all Stackdriver notifications and publish them to PubSub topic.

## What it does

Creates Stackdriver channel (WebHook) which targets Cloud Run handler service. That service validates the token shared with WebHook to make sure the notifications are from valid source and relays all messages to a PubSub topic.

## Why

Stackdriver has a limit of 16 notification channels that can be used for incidents. This service allows you to create only one WebHook and route all notifications to PubSub topic via Cloud Run service so you can create any number of Stackdriver policies that will trigger unlimited number of alerts. This way you can create multiple subscriptions to the notifications topic to handle these alerts.

## Notifications

The notification published to PubSub topic will differ in content depending the policy that triggered them (see [samples](https://cloud.google.com/monitoring/alerts/policies-in-json)) but generally they will look like this JSON message

```json
{
    "incident": {
        "incident_id": "0.lekp2pr4h14z",
        "resource_id": "",
        "resource_name": "cloudylabs Cloud Pub/Sub Subscription labels {subscription_id=pubsub-to-bigquery-pump-sub}",
        "resource": {
            "type": "pubsub_subscription",
            "labels": {
                "subscription_id": "pubsub-to-bigquery-pump-sub"
            }
        },
        "started_at": 1573487005,
        "policy_name": "stackdriver-notifs-policy",
        "condition_name": "num-undelivered-messages",
        "url": "https://app.google.stackdriver.com/incidents/0.lekp2pr4h14z?project=cloudylabs",
        "state": "open",
        "ended_at": null,
        "summary": "Unacked messages for cloudylabs Cloud Pub/Sub Subscription labels {subscription_id=pubsub-to-bigquery-pump-sub} is above the threshold of 100 with a value of 262.000."
    },
    "version": "1.2"
}
```

## Prerequisites

If you don't have one already, start by creating new project and configuring [Google Cloud SDK](https://cloud.google.com/sdk/docs/). Similarly, if you have not done so already, you will have [set up Cloud Run](https://cloud.google.com/run/docs/setup).


## Deployment

> Optional: If you want to change the name and region of the deployed service review the [bin/config](bin/config) file

> To keep this readme short, I will be asking you to execute scripts. You should review these scripts before execution.

### PubSub Topic

Execute [bin/topic](bin/topic) script to create the topic where all notifications will be published

```
bin/topic
```

## IAM Account

Execute [bin/account](bin/account) script to create IAM account which will be used to run Cloud Run service. The created account will be granted only the necessary roles (`run.invoker`, `pubsub.publisher`, `logging.logWriter`, `monitoring.metricWriter`)

```
bin/account
```

## Cloud Run Service

Execute the [bin/deploy](bin/deploy) script to create Cloud Run service that will be used to handle all Stackdriver notifications.

```
bin/deploy
```


## Stackdriver

To setup the Stackdriver side of this solution you will need to create Channel and at least one Policy.

### Channel

Stackdriver supports WebHooks to notify remote services about incidents that occur. To set up this WebHooks channel execute the [bin/channel](bin/channel) script.

```
bin/channel
```

### Policy

To monitor GCP and even AWS resources, you need to create alerting policies that when triggered, will use the above created channel (WebHook) to send notifications to Cloud Run. In the [policy/](policy/) directory you will find a few sample policies. Here is an example of policy monitoring PubSub topic for un acknowledged messages (age or number)

```yaml
---
combiner: OR
conditions:
- conditionThreshold:
    aggregations:
    - alignmentPeriod: 60s
      perSeriesAligner: ALIGN_MEAN
    comparison: COMPARISON_GT
    duration: 60s
    filter: metric.type="pubsub.googleapis.com/subscription/oldest_unacked_message_age"
      resource.type="pubsub_subscription" resource.label."subscription_id"="my-iot-events-pump"
    thresholdValue: 180
    trigger:
      count: 1
  displayName: oldest-unacked-message-age
- conditionThreshold:
    aggregations:
    - alignmentPeriod: 60s
      perSeriesAligner: ALIGN_MEAN
    comparison: COMPARISON_GT
    duration: 60s
    filter: metric.type="pubsub.googleapis.com/subscription/num_undelivered_messages"
      resource.type="pubsub_subscription" resource.label."subscription_id"="my-iot-events-pump"
    thresholdValue: 100
    trigger:
      count: 1
  displayName: num-undelivered-messages
enabled: true
```

That policy will result in this in Stackdriver

![](images/policy.png)

To create a policy you will need to exit the [bin/policy](bin/policy) script with the path of your policy file and then execute it.

```
bin/policy
```

## Cleanup

To cleanup all resources created by this sample execute

```shell
bin/cleanup
```

## Disclaimer

This is my personal project and it does not represent my employer. I take no responsibility for issues caused by this code. I do my best to ensure that everything works, but if something goes wrong, my apologies is all you will get.


