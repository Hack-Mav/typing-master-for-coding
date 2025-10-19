#!/bin/bash

# Deploy script for Google Cloud CDN setup
# This script sets up Cloud CDN for the Typing Master application

set -e

PROJECT_ID="typing-master-prod"
BUCKET_NAME="typing-master-static"
REGION="us-central1"
DOMAIN="typing-master.app"
WWW_DOMAIN="www.typing-master.app"

echo "Setting up Google Cloud CDN for Typing Master..."

# Set project
gcloud config set project $PROJECT_ID

# 1. Create Cloud Storage bucket
echo "Creating Cloud Storage bucket..."
gsutil mb -p $PROJECT_ID -c STANDARD -l $REGION gs://$BUCKET_NAME || echo "Bucket already exists"

# 2. Set bucket lifecycle (optional - clean up old versions)
cat > lifecycle.json <<EOF
{
  "lifecycle": {
    "rule": [
      {
        "action": {"type": "Delete"},
        "condition": {
          "age": 90,
          "matchesPrefix": ["old/"]
        }
      }
    ]
  }
}
EOF
gsutil lifecycle set lifecycle.json gs://$BUCKET_NAME

# 3. Make bucket publicly readable
echo "Setting bucket permissions..."
gsutil iam ch allUsers:objectViewer gs://$BUCKET_NAME

# 4. Set CORS configuration
echo "Setting CORS configuration..."
gsutil cors set cors.json gs://$BUCKET_NAME

# 5. Create backend bucket
echo "Creating backend bucket..."
gcloud compute backend-buckets create typing-master-static-backend \
  --gcs-bucket-name=$BUCKET_NAME \
  --enable-cdn \
  --cache-mode=CACHE_ALL_STATIC \
  --default-ttl=3600 \
  --max-ttl=86400 \
  --client-ttl=3600 \
  --negative-caching \
  --negative-caching-policy=404=300,500=0 || echo "Backend bucket already exists"

# 6. Create URL map
echo "Creating URL map..."
gcloud compute url-maps create typing-master-cdn \
  --default-backend-bucket=typing-master-static-backend || echo "URL map already exists"

# 7. Create managed SSL certificate
echo "Creating SSL certificate..."
gcloud compute ssl-certificates create typing-master-ssl \
  --domains=$DOMAIN,$WWW_DOMAIN \
  --global || echo "SSL certificate already exists"

# 8. Create HTTPS proxy
echo "Creating HTTPS proxy..."
gcloud compute target-https-proxies create typing-master-https-proxy \
  --url-map=typing-master-cdn \
  --ssl-certificates=typing-master-ssl \
  --global || echo "HTTPS proxy already exists"

# 9. Reserve static IP
echo "Reserving static IP..."
gcloud compute addresses create typing-master-ip \
  --ip-version=IPV4 \
  --global || echo "IP already reserved"

# Get the IP address
IP_ADDRESS=$(gcloud compute addresses describe typing-master-ip --global --format="get(address)")
echo "Static IP: $IP_ADDRESS"

# 10. Create forwarding rule
echo "Creating forwarding rule..."
gcloud compute forwarding-rules create typing-master-https-rule \
  --address=typing-master-ip \
  --global \
  --target-https-proxy=typing-master-https-proxy \
  --ports=443 || echo "Forwarding rule already exists"

# 11. Create HTTP to HTTPS redirect
echo "Creating HTTP to HTTPS redirect..."
gcloud compute url-maps import typing-master-redirect \
  --global \
  --source=/dev/stdin <<EOF
defaultUrlRedirect:
  redirectResponseCode: MOVED_PERMANENTLY_DEFAULT
  httpsRedirect: true
EOF

gcloud compute target-http-proxies create typing-master-http-proxy \
  --url-map=typing-master-redirect \
  --global || echo "HTTP proxy already exists"

gcloud compute forwarding-rules create typing-master-http-rule \
  --address=typing-master-ip \
  --global \
  --target-http-proxy=typing-master-http-proxy \
  --ports=80 || echo "HTTP forwarding rule already exists"

echo ""
echo "CDN setup complete!"
echo "Static IP: $IP_ADDRESS"
echo ""
echo "Next steps:"
echo "1. Point your DNS A records for $DOMAIN and $WWW_DOMAIN to $IP_ADDRESS"
echo "2. Wait for SSL certificate to be provisioned (can take up to 15 minutes)"
echo "3. Upload static assets to gs://$BUCKET_NAME"
echo ""
echo "Check SSL certificate status:"
echo "gcloud compute ssl-certificates describe typing-master-ssl --global"
