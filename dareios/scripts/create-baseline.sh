#!/bin/bash

# Define APIs and scenarios
apis=("alexandros" "herodotos" "sokrates" "dionysios")

# Loop through all combinations of APIs and scenarios
for api in "${apis[@]}"; do
      # Run k6 test with the current settings
      k6 run --env API_NAME=$api homeros.js \
          --summary-export "report-${api}.json"
done
