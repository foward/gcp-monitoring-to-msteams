# gcp-monitoring-to-msteams using PubSub

⚙️ A simple Google Cloud Function in Go to transform / proxy [Google Cloud Platform (GCP) Monitoring](https://cloud.google.com/monitoring) pubsub incident notifications to [Microsoft Teams](https://teams.microsoft.com/).

_Forked from github.com/courtsite/gcp-monitoring-to-teams and then I have modified the code to adjust to the PubSub channel


## Getting Started

### Prerequisites

- Ensure you have `gcloud` installed:
    - MacOS: `brew cask install google-cloud-sdk`
    - Others: https://cloud.google.com/sdk/gcloud
- Ensure you have authenticated with Google Cloud: `gcloud init`
- (Optional) Set your current working project: `gcloud config set project <project>`

### Deployment

1. Clone / download a copy of this repository
2. Declare the TEAMS_WEBHOOK_URL in the Google Cloud Function Runtime Environment Variable. 
3. Run:
```
gcloud functions deploy YOUR_FUNCTION_NAME \
--set-env-vars TEAMS_WEBHOOK_URL=https://YOUR-ORG.webhook.office.com/webhookb2/xxxxx@yyy/IncomingWebhook/xxxxxxxxxx/yyyyyyyyyyyyy \
--entry-point=ReportAlertToTeams \
--memory=256MB \
--region=YOUR_REGION \
--runtime=go118 \
--trigger-topic=YOUR_PUBSUB_TOPIC \
--min-instances 0 \
--max-instances 3 \ 
--ingress-settings internal-only
```
4. Configure GCP Monitoring PubSub notification channel in https://console.cloud.google.com/monitoring/alerting/notifications
and please dont forget to add the service account from GCP Monitoring with a publisher role in the topic that you have created.

