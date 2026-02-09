#!/bin/bash

# Load Generator for MetalMart
# Generates orders at a configurable rate to stress Kafka for Mirrord demo

RATE=${1:-10}  # Orders per second (default: 10)
DURATION=${2:-60}  # Duration in seconds (default: 60)
BASE_URL=${BASE_URL:-"http://localhost:8084"}

echo "🚀 MetalMart Load Generator"
echo "=========================="
echo "Rate: $RATE orders/second"
echo "Duration: $DURATION seconds"
echo "Total orders: $((RATE * DURATION))"
echo ""

# Product IDs to use (from catalogue)
PRODUCTS=("1" "2" "3" "4" "5" "6" "7" "8" "9" "10")

# Generate random email for filtering
generate_email() {
    local id=$1
    echo "load-test-${id}@metalbear.com"
}

# Generate random customer name
generate_name() {
    local names=("Alice" "Bob" "Charlie" "Diana" "Eve" "Frank" "Grace" "Henry")
    echo "${names[$((RANDOM % ${#names[@]}))]}"
}

# Create a single order
create_order() {
    local order_id=$1
    local product_id=${PRODUCTS[$((RANDOM % ${#PRODUCTS[@]}))]}
    local email=$(generate_email $order_id)
    local name=$(generate_name)
    
    curl -s -X POST "${BASE_URL}/api/orders" \
        -H "Content-Type: application/json" \
        -d "{
            \"customer_email\": \"${email}\",
            \"customer_name\": \"${name}\",
            \"shipping_address\": {
                \"street\": \"${order_id} Test St\",
                \"city\": \"Test City\",
                \"state\": \"TS\",
                \"zip_code\": \"12345\",
                \"country\": \"USA\"
            },
            \"items\": [
                {
                    \"product_id\": \"${product_id}\",
                    \"product_name\": \"Test Product\",
                    \"quantity\": $((RANDOM % 3 + 1)),
                    \"price\": $((RANDOM % 50 + 10)).99
                }
            ],
            \"reservation_id\": \"res_load_${order_id}\"
        }" > /dev/null
    
    echo -n "."
}

# Calculate delay between requests
DELAY=$(echo "scale=3; 1/$RATE" | bc)

echo "Starting load generation..."
echo "Press Ctrl+C to stop early"
echo ""

START_TIME=$(date +%s)
END_TIME=$((START_TIME + DURATION))
ORDER_COUNT=0

while [ $(date +%s) -lt $END_TIME ]; do
    ORDER_COUNT=$((ORDER_COUNT + 1))
    create_order $ORDER_COUNT
    
    # Sleep to maintain rate
    sleep $DELAY
done

echo ""
echo ""
echo "✅ Load generation complete!"
echo "Total orders created: $ORDER_COUNT"
echo "Average rate: $(echo "scale=2; $ORDER_COUNT/$DURATION" | bc) orders/second"
