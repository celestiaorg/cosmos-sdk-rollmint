URL=localhost:26657/block\?height=3

# Define the maximum number of retries
MAX_RETRIES=50

# Define the delay between retries in seconds
RETRY_DELAY=5

# Counter for the number of retries
RETRY_COUNT=0

# Loop until the result is not equal to the expected string or maximum retries are reached
while [[ $RETRY_COUNT -lt $MAX_RETRIES ]]; do
  # Execute the curl command and capture the result
  RESULT=$(curl -s "$URL" | jq '.result')

  # Compare the result with the expected string or null string
  if [["$RESULT" != "null"]]; then
    echo "Success! The result is now not null. Specifically, it's:"
    echo $RESULT
    exit 0
    break
  fi
  echo "i think the result is null."

  # Increment the retry count
  ((RETRY_COUNT++))

  # Display a retry message
  echo "Retrying... (Attempt $RETRY_COUNT)"

  # Sleep for the specified delay before the next retry
  sleep $RETRY_DELAY
done

# Check if maximum retries are reached without success
if [[ $RETRY_COUNT -eq $MAX_RETRIES ]]; then
  echo "Maximum retries reached. Unable to obtain a different result."
fi
exit 1
