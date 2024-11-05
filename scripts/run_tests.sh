#!/bin/bash
set -e

BOT_ID=88EDKAHRCV
BOT_ALIAS_ID=TSTALIASID
REGION=us-east-1
SCRIPT_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
PROJECT_ROOT="$( cd "$SCRIPT_DIR/.." && pwd )"
TEST_CASES_PATH="$PROJECT_ROOT/tests/test_cases.csv"

echo "Using test cases from: $TEST_CASES_PATH"

if [ ! -f "$TEST_CASES_PATH" ]; then
    echo "Error: Test cases file not found at $TEST_CASES_PATH"
    exit 1
fi

# Initialize test results
echo "[]" > test_results.json
FAILED_TESTS=0
TOTAL_TESTS=0

# Read CSV and run tests
while IFS=, read -r test_id intent utterance expected_response slots prompt_name || [ -n "$test_id" ]
do
    # Skip header
    if [ "$test_id" = "testCaseId" ]; then continue; fi

    TOTAL_TESTS=$((TOTAL_TESTS + 1))
    echo "Running test: $test_id"
    echo "Utterance: $utterance"

    # Create a session ID for this test
    SESSION_ID="test-$test_id-$(date +%s)"

    # Call Lex Runtime API
    RESPONSE=$(aws lexv2-runtime recognize-text \
        --bot-id "$BOT_ID" \
        --bot-alias-id "$BOT_ALIAS_ID" \
        --locale-id "en_US" \
        --session-id "$SESSION_ID" \
        --text "$utterance" \
        --region "$REGION" \
        --output json)

    # Extract actual intent and slots from response
    ACTUAL_INTENT=$(echo "$RESPONSE" | jq -r '.interpretations[0].intent.name // "NO_INTENT"')
    ACTUAL_MESSAGE=$(echo "$RESPONSE" | jq -r '.messages[0].content // "NO_MESSAGE"')
    ACTUAL_SLOTS=$(echo "$RESPONSE" | jq -r '.interpretations[0].intent.slots // {}')
    CONFIDENCE=$(echo "$RESPONSE" | jq -r '.interpretations[0].intent.confirmationState // "NONE"')

    # Determine if test passed
    TEST_PASSED=true
    FAILURE_REASONS=()

    # Check intent match
    if [ "$ACTUAL_INTENT" != "$intent" ]; then
        TEST_PASSED=false
        FAILURE_REASONS+=("Intent mismatch: expected '$intent', got '$ACTUAL_INTENT'")
    fi

    # Check response if provided
    if [ ! -z "$expected_response" ] && [ "$expected_response" != '""' ]; then
        if [[ ! "$ACTUAL_MESSAGE" =~ .*"$expected_response".* ]]; then
            TEST_PASSED=false
            FAILURE_REASONS+=("Response mismatch: expected '$expected_response', got '$ACTUAL_MESSAGE'")
        fi
    fi

    # Check slots if provided
    if [ ! -z "$slots" ] && [ "$slots" != '""' ]; then
        # Compare actual slots with expected slots
        if ! echo "$ACTUAL_SLOTS" | jq --arg expect "$slots" -e 'def deepEqual(a;b):
            if a == b then true
            elif (a|type) != (b|type) then false
            elif (a|type) == "object" then
                (a|keys) == (b|keys) and
                (a|keys|all(. as $k; deepEqual(a[$k]; b[$k])))
            elif (a|type) == "array" then
                (a|length) == (b|length) and
                range(0; a|length) as $i all(deepEqual(a[$i]; b[$i]))
            else false end;
            deepEqual(., ($expect|fromjson))' > /dev/null; then
            TEST_PASSED=false
            FAILURE_REASONS+=("Slot mismatch: expected $slots, got $ACTUAL_SLOTS")
        fi
    fi

    # Record test result
    TEST_RESULT=$(cat <<EOF
    {
        "testId": "$test_id",
        "passed": $TEST_PASSED,
        "utterance": "$utterance",
        "expectedIntent": "$intent",
        "actualIntent": "$ACTUAL_INTENT",
        "expectedResponse": $expected_response,
        "actualResponse": "$ACTUAL_MESSAGE",
        "expectedSlots": $slots,
        "actualSlots": $ACTUAL_SLOTS,
        "confidence": "$CONFIDENCE",
        "failureReasons": $(printf '%s\n' "${FAILURE_REASONS[@]}" | jq -R . | jq -s .)
    }
EOF
    )

    # Append to results file
    jq --argjson result "$TEST_RESULT" '. += [$result]' test_results.json > temp.json && mv temp.json test_results.json

    if [ "$TEST_PASSED" = false ]; then
        FAILED_TESTS=$((FAILED_TESTS + 1))
        echo "❌ Test failed: $test_id"
        printf '%s\n' "${FAILURE_REASONS[@]}"
    else
        echo "✅ Test passed: $test_id"
    fi

    echo "----------------------------------------"

done < "$TEST_CASES_PATH"

# Generate summary
PASSED_TESTS=$((TOTAL_TESTS - FAILED_TESTS))
echo "Test Summary:"
echo "Total Tests: $TOTAL_TESTS"
echo "Passed: $PASSED_TESTS"
echo "Failed: $FAILED_TESTS"

# Add summary to results
jq --arg total "$TOTAL_TESTS" \
   --arg passed "$PASSED_TESTS" \
   --arg failed "$FAILED_TESTS" \
   '. | {"summary": {"total": $total, "passed": $passed, "failed": $failed}, "results": .}' \
   test_results.json > final_results.json

mv final_results.json test_results.json

# Exit with failure if any tests failed
if [ "$FAILED_TESTS" -gt 0 ]; then
    exit 1
fi