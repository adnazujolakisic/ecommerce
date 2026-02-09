#!/bin/sh

# Seed inventory for all products in the catalogue

# Get all products from the catalogue service
PRODUCTS=$(curl -s http://catalogue:8081/api/products)

# Log the products to the console
echo "Products: $PRODUCTS"

# Iterate over each product and seed inventory
echo "$PRODUCTS" | jq -c '.[]' | while read -r product; do
  echo "Processing product: $product"
  PRODUCT_ID=$(echo "$product" | jq -r '.id')
  echo "Extracted PRODUCT_ID: $PRODUCT_ID"
  
  # Check if PRODUCT_ID is empty
  if [ -z "$PRODUCT_ID" ]; then
    echo "Error: Product ID is empty. Skipping..."
    continue
  fi

  # Seed inventory with a random quantity between 10 and 100
  QUANTITY=$((10 + RANDOM % 91))
  
  echo "Seeding inventory for product $PRODUCT_ID with quantity $QUANTITY"
  
  curl -s -X POST -H "Content-Type: application/json" \
    -d "{\"product_id\": \"$PRODUCT_ID\", \"quantity\": $QUANTITY}" \
    http://inventory:8082/api/inventory/init
done
