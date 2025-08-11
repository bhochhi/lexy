#!/usr/bin/env bash
set -euo pipefail

STACK=LexyStack
ACCOUNT_ID=${CDK_DEFAULT_ACCOUNT:-$(aws sts get-caller-identity --query Account --output text 2>/dev/null || echo '')}
REGION=${AWS_REGION:-${CDK_DEFAULT_REGION:-us-east-1}}
REMOVE_BOOTSTRAP=false
FORCE=false

usage() {
  cat <<EOF
Usage: $0 [options]
Safely destroy the CDK application stack (and optionally the bootstrap stack).

Options:
  --region <region>           AWS region (defaults to env AWS_REGION or us-east-1)
  --stack <name>              Stack name (default: LexyStack)
  --remove-bootstrap          ALSO delete the CDKToolkit bootstrap stack (careful: shared!)
  --force                     Skip confirmation prompts
  -h|--help                   Show this help

Examples:
  $0 --region us-east-1
  $0 --remove-bootstrap --force
EOF
}

while [[ ${1:-} =~ ^- ]]; do
  case "$1" in
    --region) shift; REGION=$1 ;;
    --stack) shift; STACK=$1 ;;
    --remove-bootstrap) REMOVE_BOOTSTRAP=true ;;
    --force) FORCE=true ;;
    -h|--help) usage; exit 0 ;;
    *) echo "Unknown option: $1" >&2; usage; exit 1 ;;
  esac
  shift || true
done

if [ -z "$ACCOUNT_ID" ]; then
  echo "Unable to determine AWS account (check credentials)." >&2
  exit 1
fi

echo "Account: $ACCOUNT_ID  Region: $REGION  Stack: $STACK"

STACK_STATUS=$(aws cloudformation describe-stacks --stack-name "$STACK" --query 'Stacks[0].StackStatus' --output text 2>/dev/null || echo 'NOT_FOUND')
if [ "$STACK_STATUS" = 'NOT_FOUND' ]; then
  echo "Stack $STACK not found (nothing to destroy)."
else
  echo "Current stack status: $STACK_STATUS"
  if [ "$FORCE" = false ]; then
    read -r -p "Destroy stack $STACK? (y/N) " ans
    [[ "$ans" =~ ^[Yy]$ ]] || { echo "Aborted."; exit 0; }
  fi
  AWS_REGION=$REGION npx cdk destroy "$STACK" --force || { echo "Destroy failed" >&2; exit 1; }
fi

if $REMOVE_BOOTSTRAP; then
  BOOT=CDKToolkit
  BOOT_STATUS=$(aws cloudformation describe-stacks --stack-name "$BOOT" --query 'Stacks[0].StackStatus' --output text 2>/dev/null || echo 'NOT_FOUND')
  if [ "$BOOT_STATUS" = 'NOT_FOUND' ]; then
    echo "Bootstrap stack $BOOT not found."
  else
    echo "Bootstrap stack status: $BOOT_STATUS"
    if [ "$FORCE" = false ]; then
      read -r -p "ALSO delete bootstrap stack $BOOT? This may affect other CDK apps. (y/N) " ans2
      [[ "$ans2" =~ ^[Yy]$ ]] || { echo "Skipped bootstrap removal."; exit 0; }
    fi
    aws cloudformation delete-stack --stack-name "$BOOT"
    echo "Waiting for bootstrap stack deletion..."
    aws cloudformation wait stack-delete-complete --stack-name "$BOOT" || true
  fi
fi

echo "Done."
