## Find the instance IP: 
## gcloud compute instances list ws-1 --project websokcet-1333 --format json | jq .[].networkInterfaces[].accessConfigs[].natIP
rsync -av -e 'ssh -i ~/.ssh/google_compute_engine' . os@130.211.84.114:/home/os/app
