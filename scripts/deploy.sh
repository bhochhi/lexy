#!/usr/bin/env bash
set -euo pipefail
STACK=LexyStack
ACCOUNT_ID=${CDK_DEFAULT_ACCOUNT:-$(aws sts get-caller-identity --query Account --output text 2>/dev/null || echo '')}
REGION=${AWS_REGION:-${CDK_DEFAULT_REGION:-us-east-1}}

if [ -z "$ACCOUNT_ID" ]; then
	echo "Could not determine AWS account (check credentials/profile)." >&2
	exit 1
fi

if [ ! -f cdk.json ]; then
	echo "cdk.json missing; creating minimal one." >&2
	echo '{"app":"node dist/bin/app.js"}' > cdk.json
fi

echo "Using account: ${ACCOUNT_ID} region: $REGION"

echo "Building project"
npm install --no-fund --no-audit
npm run build

echo "Synthesizing CDK"
AWS_REGION=$REGION npx cdk synth || { echo "Synth failed" >&2; exit 1; }

bootstrap() {
	echo "Bootstrapping environment (account=$ACCOUNT_ID region=$REGION)" >&2
	AWS_REGION=$REGION npx cdk bootstrap aws://$ACCOUNT_ID/$REGION \
		--cloudformation-execution-policies arn:aws:iam::aws:policy/AdministratorAccess "$@"
}

# Detect rollback state and clean if necessary
STATUS=$(aws cloudformation describe-stacks --stack-name CDKToolkit --query 'Stacks[0].StackStatus' --output text 2>/dev/null || echo 'NONE')
if [[ "$STATUS" == *ROLLBACK_COMPLETE ]]; then
	echo "Found bootstrap stack in $STATUS. Deleting and retrying..." >&2
	aws cloudformation delete-stack --stack-name CDKToolkit
	aws cloudformation wait stack-delete-complete --stack-name CDKToolkit || true
fi

set +e
bootstrap
RC=$?
set -e
if [ $RC -ne 0 ]; then
	echo "Initial bootstrap failed. Forcing re-bootstrap..." >&2
	bootstrap --force || { echo "Forced bootstrap failed" >&2; exit 1; }
fi

echo "Deploying stack"
AWS_REGION=$REGION npx cdk deploy ${STACK} --require-approval never || { echo "Deploy failed" >&2; exit 1; }

echo "Success. Retrieve ApiUrl output above."
