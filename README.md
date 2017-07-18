# grpcsimulator


## Development

Build:

    docker build . -t gcr.io/badgerodon-173120/gosimulator:$VERSION

Run:

    docker run \
        -p 5000:80 \
        -v $HOME/.config/gcloud/application_default_credentials.json:/root/gcloud.credentials \
        -e "GOOGLE_APPLICATION_CREDENTIALS=/root/gcloud.credentials" \
        -i \
        -t gcr.io/badgerodon-173120/gosimulator:$VERSION

Deploy:

    gcloud docker -- push gcr.io/badgerodon-173120/gosimulator:
