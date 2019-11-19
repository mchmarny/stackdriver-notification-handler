# Generic Stackdriver Alert WebHook Handler

This simple Cloud Run service handles all Stackdriver notifications resulting from alerting policies and publishes them to PubSub topic for additional handlers to process them downstream (see [pubsub-to-bigquery-pump](https://github.com/mchmarny/pubsub-to-bigquery-pump) for example)

## What

Creates a single Stackdriver channel (WebHook) which targets Cloud Run handler service. Creates one or more Alerting Policies in Stackdriver that will send notifications through the previously created channel. That Cloud Run service validates the token shared with WebHook to make sure the notifications are from a valid source, and then, relays these messages to a PubSub topic to be processed by additional handlers.

## Why

Stackdriver alerting policies can capture many interesting events in GCP that currently do not have established event triggers.

Stackdriver has also a limit of 16 notification channels that can be used for incident notifications. To rather than creating individual WebHooks for each type alert you need, his service allows you to create a single WebHook and route all policy created notifications to a PubSub topic so you can create any number of Stackdriver policies that will trigger unlimited number of alerts.

## Notifications

The notification published to PubSub topic will differ in content depending on the policy that triggered them (see [alert samples](https://cloud.google.com/monitoring/alerts/policies-in-json))). Here is an example of incident alert for metered resource (e.g. PubSub Topic)

```json
{
  "incident": {
    "incident_id": "0.lekp2pr4h14z",
    "resource_id": "",
    "resource_name": "cloudylabs Cloud Pub/Sub Subscription labels {subscription_id=pubsub-to-bigquery-pump-sub}",
    "resource": {
      "type": "pubsub_subscription",
      "labels": { "subscription_id": "pubsub-to-bigquery-pump-sub" }
    },
    "started_at": 1573487005,
    "policy_name": "stackdriver-notifs-policy",
    "condition_name": "num-undelivered-messages",
    "url": "https://app.google.stackdriver.com/incidents/0.lekp2pr4h14z?project=cloudylabs",
    "state": "open",
    "ended_at": null,
    "summary": "Unacked messages for Cloud Pub/Sub Subscription labels 'subscription_id=pubsub-to-bigquery-pump-sub' is above the threshold of 100 with a value of 262.000."
  },
  "version": "1.2"
}
```

## Prerequisites

If you don't have one already, start by creating new project and configuring [Google Cloud SDK](https://cloud.google.com/sdk/docs/). Similarly, if you have not done so already, you will have [set up Cloud Run](https://cloud.google.com/run/docs/setup).


## Deployment

### Configuration

To simplify the following commands we will first capture the project ID and notification token

```shell
export PROJECT=$(gcloud config get-value project)
echo "Project: ${PROJECT}"
export NOTIF_TOKEN=$(openssl rand -base64 32)
echo "Project: ${Token}"
```

### PubSub Topic

Create the topic (`stackdriver-notifications`) where all notifications will be published

```shell
gcloud pubsub topics create stackdriver-notifications
```

## IAM Account

Create IAM account (`sd-notif-handler`) which will be used to run Cloud Run service.

```shell
gcloud iam service-accounts create sd-notif-handler \
  --display-name "stackdriver-notification cloud run service account"
```

To allow this account to perform the necessary functions we are going to grant it a few roles

```shell
gcloud projects add-iam-policy-binding $PROJECT \
  --member "serviceAccount:sd-notif-handler@${PROJECT}.iam.gserviceaccount.com" \
  --role roles/run.invoker

# TODO: `pubsub.publisher` should be sufficient
gcloud projects add-iam-policy-binding $PROJECT \
  --member "serviceAccount:sd-notif-handler@${PROJECT}.iam.gserviceaccount.com" \
  --role roles/pubsub.editor

gcloud projects add-iam-policy-binding $PROJECT \
  --member "serviceAccount:sd-notif-handler@${PROJECT}.iam.gserviceaccount.com" \
  --role roles/logging.logWriter

gcloud projects add-iam-policy-binding $PROJECT \
  --member "serviceAccount:sd-notif-handler@${PROJECT}.iam.gserviceaccount.com" \
  --role roles/cloudtrace.agent

gcloud projects add-iam-policy-binding $PROJECT \
  --member "serviceAccount:sd-notif-handler@${PROJECT}.iam.gserviceaccount.com" \
  --role roles/monitoring.metricWriter
```

## Cloud Run Service

Create Cloud Run service that will be used to handle all Stackdriver notifications.

```shell
gcloud beta run deploy sd-notif-handler \
  --allow-unauthenticated \
  --image gcr.io/cloudylabs-public/sd-notif-handler:0.1.1 \
  --platform managed \
  --timeout 15m \
  --region us-central1 \
  --set-env-vars "RELEASE=v0.1.1,TOPIC_NAME=stackdriver-notifications,NOTIF_TOKEN=${NOTIF_TOKEN}" \
  --service-account "sd-notif-handler@${PROJECT}.iam.gserviceaccount.com"
```

Once the service is created, we are also going to add a policy binding

```shell
gcloud beta run services add-iam-policy-binding sd-notif-handler \
  --member "serviceAccount:sd-notif-handler@${PROJECT}.iam.gserviceaccount.com" \
  --role roles/run.invoker
```

## Stackdriver

With the processing service ready, we can now define the Stackdriver channel and one or more policies.

### Channel

Stackdriver supports WebHooks to notify remote services about incidents that occur. To set up this first create a WebHooks channel

```shell
export SERVICE_URL=$(gcloud beta run services describe sd-notif-handler \
  --region us-central1 --format="value(status.domain)")
echo "SERVICE_URL=${SERVICE_URL}"

gcloud alpha monitoring channels create \
  --display-name sd-notif-handler-channel \
  --channel-labels "url=${SERVICE_URL}/v1/notif?token=${NOTIF_TOKEN}" \
  --type webhook_tokenauth \
  --enabled
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

![](image/policy.png)

Once you have your policy file defined, you can create a policy and assign it to the above created channel

```shell
export CHANNEL_ID=$(gcloud alpha monitoring channels list \
  --filter "displayName='sd-notif-handler-channel'" \
  --format 'value("name")')

gcloud alpha monitoring policies create \
  --display-name sd-notif-handler-policy \
  --notification-channels $CHANNEL_ID \
  --policy-from-file PATH_TO_YOUR_POLICY_FILE.yaml
```

## Cleanup

To cleanup all resources created by this sample execute

```shell
export POLICY_ID=$(gcloud alpha monitoring policies list \
  --filter "displayName='sd-notif-handler-policy'" \
  --format 'value("name")')
gcloud alpha monitoring policies delete $POLICY_ID

export CHANNEL_ID=$(gcloud alpha monitoring channels list \
  --filter "displayName='sd-notif-handler-channel'" \
  --format 'value("name")')
gcloud alpha monitoring channels delete $CHANNEL_ID

gcloud pubsub subscriptions delete stackdriver-notifications

gcloud beta run services delete sd-notif-handler \
  --platform managed \
  --region us-central1

gcloud iam service-accounts delete \
  "sd-notif-handler@${PROJECT}.iam.gserviceaccount.com"
```

## Disclaimer

This is my personal project and it does not represent my employer. I take no responsibility for issues caused by this code. I do my best to ensure that everything works, but if something goes wrong, my apologies is all you will get.


