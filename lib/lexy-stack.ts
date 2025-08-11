import { Duration, Stack, StackProps, CfnOutput, Aws } from 'aws-cdk-lib';
import { Construct } from 'constructs';
import { Runtime } from 'aws-cdk-lib/aws-lambda';
import { NodejsFunction } from 'aws-cdk-lib/aws-lambda-nodejs';
import { RetentionDays } from 'aws-cdk-lib/aws-logs';
import { PolicyStatement, Effect, Role, ServicePrincipal, ManagedPolicy } from 'aws-cdk-lib/aws-iam';
import { RestApi, LambdaIntegration, Deployment, Stage } from 'aws-cdk-lib/aws-apigateway';
import { CfnBot, CfnBotVersion, CfnBotAlias } from 'aws-cdk-lib/aws-lex';

function lexIntent(name: string, sampleUtterances: string[], slots: any[] = []) {
  return {
    name,
    description: name,
    sampleUtterances: sampleUtterances.map(u => ({ utterance: u })),
    intentClosingSetting: {
      closingResponse: {
        messageGroupsList: [
          { message: { plainTextMessage: { value: 'Thanks, this intent is now complete.' } } }
        ]
      }
    },
    fulfillmentCodeHook: { enabled: false },
    slotPriorities: slots.map((s, idx) => ({ slotName: s.name, priority: idx + 1 })),
    slots
  };
}

function slot(name: string, type: string, prompt: string, required = true) {
  return {
    name,
    description: name,
    slotTypeName: type,
    valueElicitationSetting: {
      slotConstraint: required ? 'Required' : 'Optional',
      promptSpecification: {
        maxRetries: 2,
        messageGroupsList: [
          { message: { plainTextMessage: { value: prompt } } }
        ],
        allowInterrupt: true
      }
    }
  };
}

export class LexyStack extends Stack {
  constructor(scope: Construct, id: string, props?: StackProps) {
    super(scope, id, props);

    const lexServiceRole = new Role(this, 'LexServiceRole', {
      assumedBy: new ServicePrincipal('lex.amazonaws.com'),
      managedPolicies: [ManagedPolicy.fromAwsManagedPolicyName('service-role/AWSLambdaBasicExecutionRole')],
      description: 'Role for Lex logging'
    });

    lexServiceRole.addToPolicy(new PolicyStatement({
      effect: Effect.ALLOW,
      actions: ['logs:CreateLogGroup', 'logs:CreateLogStream', 'logs:PutLogEvents'],
      resources: ['*']
    }));

    const lexBot = new CfnBot(this, 'ChatBot', {
      dataPrivacy: {}, // will override with correct casing ChildDirected via escape hatch
      idleSessionTtlInSeconds: 600,
      name: 'LexyBot',
      roleArn: lexServiceRole.roleArn,
      botLocales: [{
        localeId: 'en_US',
        description: 'English US',
        nluConfidenceThreshold: 0.40,
        voiceSettings: { voiceId: 'Salli' },
        intents: [
          lexIntent('routing_number', ['What is the routing number?', 'Give me the routing number for checking'], [
            slot('accountType', 'AMAZON.AlphaNumeric', 'Which account type? (checking/savings)', false),
            slot('state', 'AMAZON.US_STATE', 'Which state? (optional)', false)
          ]),
          lexIntent('dispute_transaction', ['I want to dispute a transaction', "There is a charge I don't recognize"], [
            slot('transactionDate', 'AMAZON.Date', 'What is the transaction date?'),
            slot('amount', 'AMAZON.Number', 'What is the amount?'),
            slot('merchant', 'AMAZON.AlphaNumeric', 'Merchant name?'),
            slot('last4', 'AMAZON.Number', 'Last 4 of the card?'),
            slot('reason', 'AMAZON.AlphaNumeric', 'Reason for dispute?')
          ]),
          lexIntent('credit_rate_interest', ["What's my credit card interest rate?"], [
            slot('cardType', 'AMAZON.AlphaNumeric', 'Card type (credit/loan)?'),
            slot('accountId', 'AMAZON.AlphaNumeric', 'Account id? (optional)', false)
          ]),
            lexIntent('pay_bill', ['Pay my electric bill', 'Transfer $200 to my mortgage'], [
            slot('payee', 'AMAZON.AlphaNumeric', 'Who is the payee?'),
            slot('amount', 'AMAZON.Number', 'How much?'),
            slot('fromAccount', 'AMAZON.AlphaNumeric', 'From which account?'),
            slot('date', 'AMAZON.Date', 'Payment date?')
          ]),
          lexIntent('transfer_funds', ['Transfer $500 to savings', 'Move money between accounts'], [
            slot('fromAccount', 'AMAZON.AlphaNumeric', 'From account?'),
            slot('toAccount', 'AMAZON.AlphaNumeric', 'To account?'),
            slot('amount', 'AMAZON.Number', 'Amount to transfer?'),
            slot('date', 'AMAZON.Date', 'Transfer date?')
          ]),
          lexIntent('report_card_lost', ['I lost my debit card', 'Block my credit card'], [
            slot('cardType', 'AMAZON.AlphaNumeric', 'Card type?'),
            slot('last4', 'AMAZON.Number', 'Last 4 digits?'),
            slot('whenLost', 'AMAZON.Date', 'When did you lose it?')
          ]),
          lexIntent('claim_insurance', ['I need to file an auto claim', 'File a property damage claim'], [
            slot('policyNumber', 'AMAZON.AlphaNumeric', 'Policy number?'),
            slot('claimType', 'AMAZON.AlphaNumeric', 'Claim type (auto/home)?'),
            slot('incidentDate', 'AMAZON.Date', 'Incident date?'),
            slot('description', 'AMAZON.AlphaNumeric', 'Brief description?'),
            slot('location', 'AMAZON.AlphaNumeric', 'Location of incident?')
          ]),
          lexIntent('coverage_info', ["What's covered under my policy?", 'Do I have rental reimbursement?'], [
            slot('policyNumber', 'AMAZON.AlphaNumeric', 'Policy number?'),
            slot('coverageItem', 'AMAZON.AlphaNumeric', 'Coverage item?')
          ]),
          lexIntent('branch_hours_location', ['Where is the nearest branch?', 'What time does the San Antonio branch open?'], [
            slot('location', 'AMAZON.City', 'Which city or zip?'),
            slot('branchName', 'AMAZON.AlphaNumeric', 'Branch name? (optional)', false)
          ]),
          lexIntent('agent_handoff', ['I want to talk to a human', 'Schedule a callback'], [
            slot('preferredTime', 'AMAZON.Time', 'Preferred time?'),
            slot('reason', 'AMAZON.AlphaNumeric', 'Reason?'),
            slot('contactNumber', 'AMAZON.PhoneNumber', 'Contact number?')
          ]),
          {
            name: 'FallbackIntent',
            parentIntentSignature: 'AMAZON.FallbackIntent'
          }
        ]
      }]
    });

  // Override DataPrivacy child key casing for CloudFormation
  lexBot.addOverride('Properties.DataPrivacy', { ChildDirected: false });

  const botVersion = new CfnBotVersion(this, 'BotVersion', {
      botId: lexBot.attrId,
      description: 'Initial version',
      botVersionLocaleSpecification: [
        { botVersionLocaleDetails: { sourceBotVersion: 'DRAFT' }, localeId: 'en_US' }
      ]
    });
    botVersion.addDependency(lexBot);

    const botAlias = new CfnBotAlias(this, 'BotAlias', {
      botAliasName: 'prod',
      botId: lexBot.attrId,
      botVersion: botVersion.attrBotVersion
    });
    botAlias.addDependency(botVersion);

    const backend = new NodejsFunction(this, 'ChatHandler', {
      runtime: Runtime.NODEJS_18_X,
      entry: 'lambda/chat-handler.ts',
      handler: 'handler',
      memorySize: 512,
      timeout: Duration.seconds(15),
      logRetention: RetentionDays.ONE_WEEK,
      environment: {
        LEX_BOT_ID: lexBot.attrId,
        LEX_BOT_ALIAS_ID: botAlias.attrBotAliasId,
        LEX_LOCALE_ID: 'en_US',
        INTENT_CONFIDENCE_THRESHOLD: '0.6',
  BEDROCK_MODEL_ID: 'amazon.nova-micro-v1:0',
  USE_BEDROCK: 'true'
      }
    });

    backend.addToRolePolicy(new PolicyStatement({
      effect: Effect.ALLOW,
      actions: ['bedrock:InvokeModel', 'bedrock:InvokeModelWithResponseStream'],
      resources: ['*']
    }));

    backend.addToRolePolicy(new PolicyStatement({
      effect: Effect.ALLOW,
      actions: ['lex:RecognizeText','lex:StartConversation','lex:PutSession','lex:GetSession','lex:DeleteSession'],
      resources: ['*']
    }));

    const api = new RestApi(this, 'ChatApi', { restApiName: 'LexyChatApi', deploy: false });
    const chat = api.root.addResource('chat');
    chat.addMethod('POST', new LambdaIntegration(backend));
    const deployment = new Deployment(this, 'ApiDeployment', { api });
    new Stage(this, 'ProdStage', { deployment, stageName: 'prod' });

    new CfnOutput(this, 'ApiUrl', { value: `https://${api.restApiId}.execute-api.${Aws.REGION}.amazonaws.com/prod/chat` });
    new CfnOutput(this, 'LexBotId', { value: lexBot.attrId });
    new CfnOutput(this, 'LexBotAliasId', { value: botAlias.attrBotAliasId });
  }
}
