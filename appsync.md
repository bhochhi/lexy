AppSync Lambda DynamoDB Integration Setup

ok i need to create AppSync service with dynamodb. Then I want to create lambda and from lambda source code, I want to call Appsync grapql for mutation. tell me what I need to do? or can you create all the necessary services needed?

I'll help you create an AppSync service with DynamoDB and a Lambda function that can call AppSync GraphQL mutations. Let me get the detailed information you need to set this up.

Retrieving knowledge

Retrieving knowledge

Based on the information I've gathered, I'll provide you with a comprehensive guide to create an AppSync service with DynamoDB and a Lambda function that can call AppSync GraphQL mutations. Here's what you need to do:

Step 1: Create AppSync GraphQL API with DynamoDB
1.1 Create the AppSync API
Sign in to the AWS Management Console and open the AppSync console 
Choose Create API
Select Design from scratch and choose Next
Name your API (e.g., "MyAppSyncAPI") and choose Next
Choose Create API
1.2 Define GraphQL Schema
Create a schema for your API. Here's an example schema:

schema {
  query: Query
  mutation: Mutation
}

type Query {
  getPost(id: ID!): Post
  listPosts: [Post]
}

type Mutation {
  createPost(input: CreatePostInput!): Post
  updatePost(input: UpdatePostInput!): Post
  deletePost(id: ID!): Post
}

type Post {
  id: ID!
  title: String!
  content: String!
  author: String!
  createdAt: String
  updatedAt: String
}

input CreatePostInput {
  title: String!
  content: String!
  author: String!
}

input UpdatePostInput {
  id: ID!
  title: String
  content: String
  author: String
}

1.3 Create DynamoDB Table
Go to the DynamoDB console 
Choose Create table
Set table name (e.g., "Posts")
Set partition key as "id" (String)
Choose Create table
1.4 Connect DynamoDB as Data Source
In your AppSync API, go to Data sources
Choose Create data source
Select Amazon DynamoDB table
Choose your DynamoDB table
Create or select an IAM role with DynamoDB permissions
1.5 Create Resolvers
Create resolvers for your mutations and queries to connect them to the DynamoDB data source.

Step 2: Create Lambda Function to Call AppSync
2.1 Create Lambda Function
Go to the Lambda console 
Choose Create function
Select Author from scratch
Name your function (e.g., "AppSyncCaller")
Choose your runtime (Node.js, Python, etc.)
2.2 Lambda Function Code Example (Node.js)
const AWS = require('aws-sdk');
const https = require('https');
const crypto = require('crypto');

const appsync = new AWS.AppSync();
const region = process.env.AWS_REGION;
const appsyncUrl = process.env.APPSYNC_URL; // Your AppSync GraphQL endpoint
const apiKey = process.env.API_KEY; // If using API Key auth

exports.handler = async (event) => {
    const mutation = `
        mutation CreatePost($input: CreatePostInput!) {
            createPost(input: $input) {
                id
                title
                content
                author
                createdAt
            }
        }
    `;
    
    const variables = {
        input: {
            title: "Sample Post",
            content: "This is a sample post created from Lambda",
            author: "Lambda Function"
        }
    };
    
    const requestBody = {
        query: mutation,
        variables: variables
    };
    
    try {
        const response = await makeGraphQLRequest(appsyncUrl, requestBody, apiKey);
        console.log('GraphQL Response:', response);
        
        return {
            statusCode: 200,
            body: JSON.stringify(response)
        };
    } catch (error) {
        console.error('Error calling AppSync:', error);
        return {
            statusCode: 500,
            body: JSON.stringify({ error: error.message })
        };
    }
};

function makeGraphQLRequest(url, body, apiKey) {
    return new Promise((resolve, reject) => {
        const postData = JSON.stringify(body);
        
        const options = {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json',
                'Content-Length': Buffer.byteLength(postData),
                'x-api-key': apiKey
            }
        };
        
        const req = https.request(url, options, (res) => {
            let data = '';
            
            res.on('data', (chunk) => {
                data += chunk;
            });
            
            res.on('end', () => {
                try {
                    const response = JSON.parse(data);
                    resolve(response);
                } catch (error) {
                    reject(error);
                }
            });
        });
        
        req.on('error', (error) => {
            reject(error);
        });
        
        req.write(postData);
        req.end();
    });
}

2.3 Lambda Function Code Example (Python)
import json
import boto3
import requests
import os

def lambda_handler(event, context):
    appsync_url = os.environ['APPSYNC_URL']
    api_key = os.environ['API_KEY']
    
    mutation = """
        mutation CreatePost($input: CreatePostInput!) {
            createPost(input: $input) {
                id
                title
                content
                author
                createdAt
            }
        }
    """
    
    variables = {
        "input": {
            "title": "Sample Post from Lambda",
            "content": "This is a sample post created from Lambda",
            "author": "Lambda Function"
        }
    }
    
    headers = {
        'Content-Type': 'application/json',
        'x-api-key': api_key
    }
    
    payload = {
        'query': mutation,
        'variables': variables
    }
    
    try:
        response = requests.post(appsync_url, json=payload, headers=headers)
        response.raise_for_status()
        
        return {
            'statusCode': 200,
            'body': json.dumps(response.json())
        }
    except Exception as e:
        print(f"Error calling AppSync: {str(e)}")
        return {
            'statusCode': 500,
            'body': json.dumps({'error': str(e)})
        }

Step 3: Configure IAM Permissions
3.1 Lambda Execution Role
Your Lambda function needs permissions to:

Write to CloudWatch Logs
Call AppSync (if using IAM auth instead of API Key)
3.2 AppSync Service Role
Your AppSync API needs permissions to:

Read/Write to DynamoDB table
Step 4: Environment Variables
Set these environment variables in your Lambda function:

APPSYNC_URL: Your AppSync GraphQL endpoint URL
API_KEY: Your AppSync API key (if using API Key authentication)
Step 5: Authentication Options
You have several authentication options:

API Key (simplest for testing)
IAM (for service-to-service calls)
Cognito User Pools
Lambda Authorizer
For Lambda calling AppSync, API Key or IAM are most common.

Step 6: Test Your Setup
Deploy your Lambda function
Test the Lambda function using the Lambda console test feature 
Check your DynamoDB table to verify data was created
Use the AppSync console query editor  to test your GraphQL API directly
Alternative: Infrastructure as Code
For a more automated approach, you can use AWS CDK, CloudFormation, or SAM to deploy all these resources together. Here's a basic CDK example structure:

// CDK example (TypeScript)
const api = new appsync.GraphqlApi(this, 'Api', {
  name: 'MyAppSyncAPI',
  schema: appsync.Schema.fromAsset('schema.graphql'),
  authorizationConfig: {
    defaultAuthorization: {
      authorizationType: appsync.AuthorizationType.API_KEY,
    },
  },
});

const table = new dynamodb.Table(this, 'PostsTable', {
  partitionKey: { name: 'id', type: dynamodb.AttributeType.STRING },
});

const dataSource = api.addDynamoDbDataSource('PostsDataSource', table);

This setup will give you a complete AppSync API with DynamoDB backend and a Lambda function that can call GraphQL mutations. The Lambda function can be triggered by various AWS services or called directly to perform operations on your GraphQL API.

Sources
Creating a simple post application using DynamoDB resolvers - AWS AppSync GraphQL 

Creating a simple post application using DynamoDB JavaScript resolvers - AWS AppSync 

Getting started: Creating your first GraphQL API in AWS AppSync - AWS AppSync GraphQL 

AWS::Serverless::GraphQLApi - AWS Serverless Application Model 

VTL resolver tutorials for AWS AppSync - AWS AppSync GraphQL 

Using an AWS AppSync API with the AWS CDK - AWS AppSync GraphQL 

Structuring a GraphQL API (blank or imported APIs) - AWS AppSync 

Building a GraphQL API with AWS AppSync Using Direct Lambda Resolvers in .NET | .NET on AWS Blog 

Launching a schema in the AWS AppSync console - AWS AppSync GraphQL 

Using HTTP resolvers in AWS AppSync - AWS AppSync GraphQL 

Using AWS Lambda resolvers in AWS AppSync - AWS AppSync GraphQL 

Using AWS Lambda resolvers in AWS AppSync - AWS AppSync GraphQL 

GraphQL API - Powertools for AWS Lambda (Python) 

Configuring authorization and authentication to secure your GraphQL APIs - AWS AppSync GraphQL 

GraphQL API - Powertools for AWS Lambda (Python) 

GraphQL API - Powertools for AWS Lambda (Python) 

GraphQL API - Powertools for AWS Lambda (Python) 

GraphQL API - Powertools for AWS Lambda (Python) 

GraphQL API - Powertools for AWS Lambda (Python) 

Code Reference Log


where do i create this Create Resolvers? can you please guide?

I'll guide you through creating resolvers in AWS AppSync step by step. Resolvers are what connect your GraphQL schema fields to your data sources (like DynamoDB).

Where to Create Resolvers
Resolvers are created in the AWS AppSync Console under your GraphQL API. Here's the detailed process:

Step-by-Step Guide to Create Resolvers
1. Navigate to Resolvers Section
Go to the AppSync Console 
Click on your GraphQL API
In the left sidebar, click on Schema
You'll see your schema with Attach buttons next to each field that needs a resolver
2. Create Resolver for createPost Mutation
Find the createPost field in your Mutation type

Click the Attach button next to createPost(input: CreatePostInput!): Post

You'll see the Create Resolver page with these options:

Data source name: Select your DynamoDB data source

Configure the request mapping template:

{
    "version": "2017-02-28",
    "operation": "PutItem",
    "key": {
        "id": $util.dynamodb.toDynamoDBJson($util.autoId())
    },
    "attributeValues": {
        "title": $util.dynamodb.toDynamoDBJson($ctx.args.input.title),
        "content": $util.dynamodb.toDynamoDBJson($ctx.args.input.content),
        "author": $util.dynamodb.toDynamoDBJson($ctx.args.input.author),
        "createdAt": $util.dynamodb.toDynamoDBJson($util.time.nowISO8601()),
        "updatedAt": $util.dynamodb.toDynamoDBJson($util.time.nowISO8601())
    }
}
Configure the response mapping template:

$util.toJson($ctx.result)
Click Save Resolver

3. Create Resolver for getPost Query
Find the getPost field in your Query type

Click Attach next to getPost(id: ID!): Post

Select your DynamoDB data source

Request mapping template:

{
    "version": "2017-02-28",
    "operation": "GetItem",
    "key": {
        "id": $util.dynamodb.toDynamoDBJson($ctx.args.id)
    }
}
Response mapping template:

$util.toJson($ctx.result)
Click Save Resolver

4. Create Resolver for listPosts Query
Find the listPosts field in your Query type

Click Attach next to listPosts: [Post]

Select your DynamoDB data source

Request mapping template:

{
    "version": "2017-02-28",
    "operation": "Scan"
}
Response mapping template:

$util.toJson($ctx.result.items)
Click Save Resolver

5. Create Resolver for updatePost Mutation
Find the updatePost field in your Mutation type

Click Attach next to updatePost(input: UpdatePostInput!): Post

Select your DynamoDB data source

Request mapping template:

{
    "version": "2017-02-28",
    "operation": "UpdateItem",
    "key": {
        "id": $util.dynamodb.toDynamoDBJson($ctx.args.input.id)
    },
    "update": {
        "expression": "SET #title = :title, #content = :content, #author = :author, #updatedAt = :updatedAt",
        "expressionNames": {
            "#title": "title",
            "#content": "content",
            "#author": "author",
            "#updatedAt": "updatedAt"
        },
        "expressionValues": {
            ":title": $util.dynamodb.toDynamoDBJson($ctx.args.input.title),
            ":content": $util.dynamodb.toDynamoDBJson($ctx.args.input.content),
            ":author": $util.dynamodb.toDynamoDBJson($ctx.args.input.author),
            ":updatedAt": $util.dynamodb.toDynamoDBJson($util.time.nowISO8601())
        }
    }
}
Response mapping template:

$util.toJson($ctx.result)
Click Save Resolver

6. Create Resolver for deletePost Mutation
Find the deletePost field in your Mutation type

Click Attach next to deletePost(id: ID!): Post

Select your DynamoDB data source

Request mapping template:

{
    "version": "2017-02-28",
    "operation": "DeleteItem",
    "key": {
        "id": $util.dynamodb.toDynamoDBJson($ctx.args.id)
    }
}
Response mapping template:

$util.toJson($ctx.result)
Click Save Resolver

Alternative: Using JavaScript Resolvers (Newer Approach)
AppSync also supports JavaScript resolvers which are easier to write. Here's how to create a JavaScript resolver:

For createPost Mutation:
When creating the resolver, select JavaScript as the runtime
Request Function:
import { util } from '@aws-appsync/utils';

export function request(ctx) {
    const { input } = ctx.args;
    const id = util.autoId();
    const now = util.time.nowISO8601();
    
    return {
        operation: 'PutItem',
        key: util.dynamodb.toMapValues({ id }),
        attributeValues: util.dynamodb.toMapValues({
            ...input,
            id,
            createdAt: now,
            updatedAt: now
        })
    };
}

export function response(ctx) {
    return ctx.result;
}
Visual Guide - Where to Find Resolvers
AppSync Console → Your API → Schema
├── Query
│   ├── getPost(id: ID!): Post [Attach] ← Click here
│   └── listPosts: [Post] [Attach] ← Click here
├── Mutation
│   ├── createPost(input: CreatePostInput!): Post [Attach] ← Click here
│   ├── updatePost(input: UpdatePostInput!): Post [Attach] ← Click here
│   └── deletePost(id: ID!): Post [Attach] ← Click here

Testing Your Resolvers
After creating all resolvers:

Go to Queries in the left sidebar of your AppSync console
Test your mutations and queries:
# Test createPost
mutation {
  createPost(input: {
    title: "My First Post"
    content: "This is the content"
    author: "John Doe"
  }) {
    id
    title
    content
    author
    createdAt
  }
}

# Test getPost
query {
  getPost(id: "your-post-id-here") {
    id
    title
    content
    author
  }
}

# Test listPosts
query {
  listPosts {
    id
    title
    author
    createdAt
  }
}

This completes the resolver setup! Your AppSync API will now be able to perform CRUD operations on your DynamoDB table, and your Lambda function can call these GraphQL operations.



so, our requitements is to have appsync, where our UI app chatbot can subscribe to some id, then we have lambda in the same VPC calls the appsync endpoint for mutabtion, the client UI chatbot get updates. for this do we even need dynamodb? help me understand ? when I create the AppSync API, it asking me for Event API or GraphQL. do I need to create both? feels like that right? then if I selected GraphQL, it has Private API configuration. help me understand and give me step by step guide.

Great question! Let me help you understand the architecture for your real-time chatbot use case and guide you through the setup.

Understanding Your Requirements
For your chatbot scenario where:

UI app subscribes to updates by ID
Lambda (in VPC) sends mutations to AppSync
Client gets real-time updates via subscriptions
Do you need DynamoDB?

Not necessarily! If you only need real-time messaging without persistence, you can use AppSync subscriptions with Local/None data sources
Yes, if you want to: Store chat history, maintain conversation state, or have message persistence
Event API vs GraphQL API:

GraphQL API - This is what you need for your use case (queries, mutations, subscriptions)
Event API - This is for simple pub/sub messaging (newer feature, simpler but less flexible)
For your chatbot, GraphQL API is the right choice as it gives you more control over data structure and real-time subscriptions.

Step-by-Step Guide
Step 1: Create AppSync GraphQL API
Go to AppSync Console 
Click Create API
Select GraphQL APIs (not Event API)
Choose Design from scratch
Click Next
Step 2: Configure API Settings
API name: ChatbotAPI
Private API configuration:
Leave unchecked for now - Private APIs are for VPC-only access
Since your UI needs to access it from browsers, use public API with proper authentication
Click Next
Skip additional configurations and click Create API
Step 3: Define GraphQL Schema (Without DynamoDB)
In the Schema section, replace the default schema with:

schema {
  query: Query
  mutation: Mutation
  subscription: Subscription
}

type Query {
  # Placeholder query (required by GraphQL)
  getStatus: String
}

type Mutation {
  # Send message to a specific chat/conversation
  sendMessage(input: MessageInput!): Message
  
  # Update typing status
  updateTypingStatus(chatId: ID!, isTyping: Boolean!): TypingStatus
}

type Subscription {
  # Subscribe to messages for a specific chat
  onMessageReceived(chatId: ID!): Message
    @aws_subscribe(mutations: ["sendMessage"])
  
  # Subscribe to typing status updates
  onTypingStatusChanged(chatId: ID!): TypingStatus
    @aws_subscribe(mutations: ["updateTypingStatus"])
}

type Message {
  id: ID!
  chatId: ID!
  content: String!
  sender: String!
  timestamp: String!
  messageType: MessageType
}

type TypingStatus {
  chatId: ID!
  userId: String!
  isTyping: Boolean!
  timestamp: String!
}

input MessageInput {
  chatId: ID!
  content: String!
  sender: String!
  messageType: MessageType
}

enum MessageType {
  USER
  BOT
  SYSTEM
}

Step 4: Create Data Sources
Since you don't need persistence, create None/Local data sources:

Go to Data sources in your AppSync API
Click Create data source
Data source name: LocalDataSource
Data source type: Select None (Local)
Click Create
Step 5: Create Resolvers
For sendMessage Mutation:
Go to Schema → Find sendMessage → Click Attach
Select LocalDataSource
Request mapping template:
{
    "version": "2017-02-28",
    "payload": {
        "id": "$util.autoId()",
        "chatId": "$ctx.args.input.chatId",
        "content": "$ctx.args.input.content",
        "sender": "$ctx.args.input.sender",
        "messageType": "$ctx.args.input.messageType",
        "timestamp": "$util.time.nowISO8601()"
    }
}

Response mapping template:
$util.toJson($ctx.result)

For updateTypingStatus Mutation:
Find updateTypingStatus → Click Attach
Select LocalDataSource
Request mapping template:
{
    "version": "2017-02-28",
    "payload": {
        "chatId": "$ctx.args.chatId",
        "userId": "$ctx.identity.sub",
        "isTyping": $ctx.args.isTyping,
        "timestamp": "$util.time.nowISO8601()"
    }
}

Response mapping template:
$util.toJson($ctx.result)

For getStatus Query:
Find getStatus → Click Attach
Select LocalDataSource
Request mapping template:
{
    "version": "2017-02-28",
    "payload": "API is running"
}

Response mapping template:
"$ctx.result"

Step 6: Configure Authentication
Go to Settings in your AppSync API
Default authorization mode: Choose based on your needs:
API Key (simplest for testing)
Amazon Cognito User Pool (for authenticated users)
AWS IAM (for service-to-service)
For a chatbot, Cognito User Pool is recommended for user authentication.

Step 7: Lambda Function for VPC
Create your Lambda function that will call AppSync mutations:

const AWS = require('aws-sdk');
const https = require('https');

const APPSYNC_URL = process.env.APPSYNC_URL;
const API_KEY = process.env.API_KEY;

exports.handler = async (event) => {
    // Process the user message and generate bot response
    const { chatId, userMessage } = event;
    
    // Your bot logic here
    const botResponse = await generateBotResponse(userMessage);
    
    // Send bot message via AppSync mutation
    const mutation = `
        mutation SendMessage($input: MessageInput!) {
            sendMessage(input: $input) {
                id
                chatId
                content
                sender
                timestamp
                messageType
            }
        }
    `;
    
    const variables = {
        input: {
            chatId: chatId,
            content: botResponse,
            sender: "ChatBot",
            messageType: "BOT"
        }
    };
    
    try {
        const response = await callAppSync(mutation, variables);
        console.log('Message sent:', response);
        
        return {
            statusCode: 200,
            body: JSON.stringify(response)
        };
    } catch (error) {
        console.error('Error:', error);
        return {
            statusCode: 500,
            body: JSON.stringify({ error: error.message })
        };
    }
};

async function callAppSync(query, variables) {
    const requestBody = JSON.stringify({ query, variables });
    
    const options = {
        method: 'POST',
        headers: {
            'Content-Type': 'application/json',
            'x-api-key': API_KEY
        }
    };
    
    return new Promise((resolve, reject) => {
        const req = https.request(APPSYNC_URL, options, (res) => {
            let data = '';
            res.on('data', chunk => data += chunk);
            res.on('end', () => {
                try {
                    resolve(JSON.parse(data));
                } catch (e) {
                    reject(e);
                }
            });
        });
        
        req.on('error', reject);
        req.write(requestBody);
        req.end();
    });
}

async function generateBotResponse(message) {
    // Your bot logic - could call other AWS services, APIs, etc.
    return `Bot response to: ${message}`;
}

Step 8: Client-Side Subscription (JavaScript)
import { API, graphqlOperation } from 'aws-amplify';

// Subscribe to messages for a specific chat
const subscription = `
    subscription OnMessageReceived($chatId: ID!) {
        onMessageReceived(chatId: $chatId) {
            id
            chatId
            content
            sender
            timestamp
            messageType
        }
    }
`;

// Subscribe to messages
const subscriptionClient = API.graphql(
    graphqlOperation(subscription, { chatId: 'your-chat-id' })
).subscribe({
    next: ({ value }) => {
        console.log('New message:', value.data.onMessageReceived);
        // Update your UI with the new message
        displayMessage(value.data.onMessageReceived);
    },
    error: (err) => console.error('Subscription error:', err)
});

// Send a message (from UI)
const sendMessage = async (chatId, content) => {
    const mutation = `
        mutation SendMessage($input: MessageInput!) {
            sendMessage(input: $input) {
                id
                chatId
                content
                sender
                timestamp
                messageType
            }
        }
    `;
    
    try {
        const result = await API.graphql(graphqlOperation(mutation, {
            input: {
                chatId,
                content,
                sender: 'User',
                messageType: 'USER'
            }
        }));
        
        console.log('Message sent:', result);
    } catch (error) {
        console.error('Error sending message:', error);
    }
};

Step 9: VPC Configuration (If Needed)
If your Lambda truly needs to be in a VPC:

Lambda VPC Configuration: Configure your Lambda with VPC, subnets, and security groups
NAT Gateway: Ensure your Lambda can reach the internet (for AppSync calls) via NAT Gateway
Security Groups: Allow outbound HTTPS (443) traffic
Step 10: Test Your Setup
Test in AppSync Console:

Go to Queries tab
Test mutations and subscriptions
Test Real-time Flow:

Client subscribes to onMessageReceived
Lambda sends mutation to sendMessage
Client receives real-time update
Architecture Summary
[UI Client] ←→ [AppSync GraphQL API] ←→ [Lambda in VPC]
     ↑              ↑                        ↑
 Subscription   Real-time Pub/Sub      Bot Processing

This setup gives you real-time messaging without requiring DynamoDB, perfect for a chatbot that doesn't need message persistence. If you later want to add message history, you can easily add DynamoDB as an additional data source.



well I need to make sure this is private API, but tell me if i just need to send data to UI but not other way, do I still need appsync to be public? all I want to ui client subscribe to appsync and after that one way flow of any data from appsync to ui as lambda sends them via appsync mutation call.

Great question! Let me clarify the architecture options for your one-way data flow scenario.

Understanding Your Use Case
You want:

UI client subscribes to AppSync
Lambda sends data via AppSync mutations
One-way flow: AppSync → UI (no UI → AppSync mutations)
Can AppSync be Private?
Short Answer: If your UI is a web browser application, AppSync cannot be fully private because browsers cannot access VPC-private resources directly.

However, you have several architectural options:

Option 1: Public AppSync with Secure Authentication (Recommended)
Architecture:

[Browser UI] ←subscription→ [Public AppSync] ←mutation← [Lambda in VPC]

Benefits:

Simplest setup
Direct browser connection to AppSync
Secure with proper authentication
Security:

Use Cognito User Pools or OIDC for authentication
Configure fine-grained authorization rules
Use API Keys with domain restrictions if needed
Setup:

Create public AppSync API
Configure authentication (Cognito recommended)
Lambda calls AppSync via internet (through NAT Gateway if in VPC)
Option 2: Private AppSync with API Gateway Proxy
Architecture:

[Browser UI] ←WebSocket→ [API Gateway] ←→ [VPC Endpoint] ←→ [Private AppSync] ←← [Lambda in VPC]

When to use:

Strict security requirements
Need VPC-only access to AppSync
More complex but fully private
Setup:

Create private AppSync API
Create VPC Endpoint for AppSync
Create API Gateway WebSocket API as proxy
Browser connects to API Gateway, which forwards to private AppSync
Option 3: Alternative Real-time Solutions
If you don't need GraphQL features, consider simpler alternatives:

A. API Gateway WebSocket
[Browser UI] ←WebSocket→ [API Gateway WebSocket] ←trigger← [Lambda]

B. Server-Sent Events (SSE)
[Browser UI] ←SSE→ [CloudFront] ←→ [API Gateway] ←→ [Lambda]

Recommended Solution: Public AppSync with Secure Auth
For your use case, I recommend Option 1 (Public AppSync) because:

Simpler architecture
Better performance (direct connection)
Cost-effective
Secure when properly configured
Step-by-Step Implementation
Step 1: Create Public AppSync API
# AppSync will be publicly accessible but secured with authentication

Run in CloudShell
Go to AppSync Console 
Create GraphQL API
Do NOT check "Private API"
Configure authentication
Step 2: Configure Authentication
Option A: Cognito User Pools (Recommended for user-based auth)

// Client configuration
import { Amplify } from 'aws-amplify';

Amplify.configure({
  aws_appsync_graphqlEndpoint: 'https://your-api.appsync-api.region.amazonaws.com/graphql',
  aws_appsync_region: 'us-east-1',
  aws_appsync_authenticationType: 'AMAZON_COGNITO_USER_POOLS',
  aws_user_pools_id: 'us-east-1_xxxxxxx',
  aws_user_pools_web_client_id: 'xxxxxxxxx',
});

Option B: API Key (Simpler for testing)

Amplify.configure({
  aws_appsync_graphqlEndpoint: 'https://your-api.appsync-api.region.amazonaws.com/graphql',
  aws_appsync_region: 'us-east-1',
  aws_appsync_authenticationType: 'API_KEY',
  aws_appsync_apiKey: 'da2-xxxxxxxxx',
});

Step 3: Schema for One-Way Flow
schema {
  query: Query
  mutation: Mutation
  subscription: Subscription
}

type Query {
  getStatus: String
}

type Mutation {
  # Only Lambda will call this
  publishUpdate(input: UpdateInput!): Update
}

type Subscription {
  # UI subscribes to this
  onUpdateReceived(channelId: ID!): Update
    @aws_subscribe(mutations: ["publishUpdate"])
}

type Update {
  id: ID!
  channelId: ID!
  message: String!
  data: AWSJSON
  timestamp: String!
  sender: String!
}

input UpdateInput {
  channelId: ID!
  message: String!
  data: AWSJSON
  sender: String!
}

Step 4: Client-Side Subscription (Read-Only)
import { API, graphqlOperation } from 'aws-amplify';

class ChatbotClient {
  constructor(channelId) {
    this.channelId = channelId;
    this.subscription = null;
  }

  // Subscribe to updates (one-way: AppSync → UI)
  subscribe(onMessage) {
    const subscription = `
      subscription OnUpdateReceived($channelId: ID!) {
        onUpdateReceived(channelId: $channelId) {
          id
          channelId
          message
          data
          timestamp
          sender
        }
      }
    `;

    this.subscription = API.graphql(
      graphqlOperation(subscription, { channelId: this.channelId })
    ).subscribe({
      next: ({ value }) => {
        const update = value.data.onUpdateReceived;
        onMessage(update);
      },
      error: (err) => {
        console.error('Subscription error:', err);
      }
    });

    return this.subscription;
  }

  // Cleanup
  unsubscribe() {
    if (this.subscription) {
      this.subscription.unsubscribe();
    }
  }
}

// Usage
const client = new ChatbotClient('user-123');

client.subscribe((update) => {
  console.log('Received update:', update);
  // Update your UI
  displayMessage(update.message, update.data);
});

Step 5: Lambda Function (Sender)
const AWS = require('aws-sdk');
const https = require('https');

const APPSYNC_URL = process.env.APPSYNC_URL;
const API_KEY = process.env.API_KEY;

exports.handler = async (event) => {
  const { channelId, message, data } = event;

  const mutation = `
    mutation PublishUpdate($input: UpdateInput!) {
      publishUpdate(input: $input) {
        id
        channelId
        message
        data
        timestamp
        sender
      }
    }
  `;

  const variables = {
    input: {
      channelId,
      message,
      data: JSON.stringify(data),
      sender: 'System'
    }
  };

  try {
    const response = await callAppSync(mutation, variables);
    console.log('Update published:', response);
    return { statusCode: 200, body: JSON.stringify(response) };
  } catch (error) {
    console.error('Error:', error);
    return { statusCode: 500, body: JSON.stringify({ error: error.message }) };
  }
};

async function callAppSync(query, variables) {
  const requestBody = JSON.stringify({ query, variables });
  
  const options = {
    method: 'POST',
    headers: {
      'Content-Type': 'application/json',
      'x-api-key': API_KEY
    }
  };

  return new Promise((resolve, reject) => {
    const req = https.request(APPSYNC_URL, options, (res) => {
      let data = '';
      res.on('data', chunk => data += chunk);
      res.on('end', () => {
        try {
          resolve(JSON.parse(data));
        } catch (e) {
          reject(e);
        }
      });
    });

    req.on('error', reject);
    req.write(requestBody);
    req.end();
  });
}

Step 6: Security Considerations
Authentication: Use Cognito or OIDC
Authorization: Add field-level auth rules
Rate Limiting: Configure in AppSync
CORS: Configure for your domain
API Key Rotation: If using API keys
# Example with auth rules
type Subscription {
  onUpdateReceived(channelId: ID!): Update
    @aws_subscribe(mutations: ["publishUpdate"])
    @aws_cognito_user_pools(cognito_groups: ["Users"])
}

type Mutation {
  publishUpdate(input: UpdateInput!): Update
    @aws_iam  # Only IAM (Lambda) can call this
}

Summary
For your one-way data flow:

Use Public AppSync with proper authentication
UI only subscribes (no mutations from UI)
Lambda sends mutations to trigger subscriptions
Secure with Cognito/OIDC for user authentication
Use IAM for Lambda → AppSync calls
This gives you the real-time capabilities you need while maintaining security and simplicity.



well in our company they don't want any public traffic directly call public appsync api. everything need to go through USAA network routes and utimately reach to gloo gateway to aws api gateway.. so what is best approach, the infrastructure team do not want you use athentication at your public appsync api.

Ah, I understand now! Your company has a strict network security policy where:

No direct public traffic to AWS services
All traffic must go through USAA network routes
Must go through Gloo Gateway → API Gateway → AWS services
No authentication at AppSync level (handled upstream)
This is a common enterprise security pattern. Here's the best approach:

Recommended Architecture
[Browser UI] → [USAA Network] → [Gloo Gateway] → [API Gateway] → [Private AppSync] ← [Lambda in VPC]

Step-by-Step Implementation
Step 1: Create Private AppSync API
Go to AppSync Console 
Create GraphQL API
Check "Private API configuration" ✅
Authentication: Set to AWS_IAM (API Gateway will handle this)
No additional auth since Gloo Gateway handles authentication
Step 2: Create VPC Endpoint for AppSync
# You'll need a VPC Endpoint to access private AppSync

Run in CloudShell
Go to VPC Console 
Endpoints → Create Endpoint
Service: com.amazonaws.region.appsync-api
VPC: Select your VPC
Subnets: Select private subnets
Security Groups: Allow HTTPS (443) from API Gateway
Step 3: Create API Gateway as Proxy
Create API Gateway that proxies GraphQL requests to private AppSync:

REST API Gateway Configuration:
# CloudFormation/SAM template
Resources:
  AppSyncProxyApi:
    Type: AWS::ApiGateway::RestApi
    Properties:
      Name: AppSyncProxy
      Description: Proxy to private AppSync
      EndpointConfiguration:
        Types:
          - REGIONAL

  AppSyncProxyResource:
    Type: AWS::ApiGateway::Resource
    Properties:
      RestApiId: !Ref AppSyncProxyApi
      ParentId: !GetAtt AppSyncProxyApi.RootResourceId
      PathPart: graphql

  AppSyncProxyMethod:
    Type: AWS::ApiGateway::Method
    Properties:
      RestApiId: !Ref AppSyncProxyApi
      ResourceId: !Ref AppSyncProxyResource
      HttpMethod: POST
      AuthorizationType: AWS_IAM
      Integration:
        Type: HTTP_PROXY
        IntegrationHttpMethod: POST
        Uri: !Sub 'https://${PrivateAppSyncApi}.appsync-api.${AWS::Region}.amazonaws.com/graphql'
        IntegrationResponses:
          - StatusCode: 200
        RequestParameters:
          integration.request.header.Authorization: context.identity.sourceIp

Step 4: WebSocket API Gateway for Subscriptions
Since subscriptions need WebSocket, create a WebSocket API:

// Lambda function for WebSocket API Gateway
const AWS = require('aws-sdk');
const appsync = new AWS.AppSync();

exports.handler = async (event) => {
    const { requestContext, body } = event;
    const { routeKey, connectionId } = requestContext;
    
    switch (routeKey) {
        case '$connect':
            return handleConnect(connectionId);
        case '$disconnect':
            return handleDisconnect(connectionId);
        case 'subscribe':
            return handleSubscribe(connectionId, JSON.parse(body));
        default:
            return { statusCode: 400, body: 'Unknown route' };
    }
};

async function handleSubscribe(connectionId, payload) {
    // Store subscription info in DynamoDB
    // Forward subscription to private AppSync
    // Set up real-time forwarding
    
    const params = {
        // AppSync subscription parameters
        query: payload.query,
        variables: payload.variables
    };
    
    // Call private AppSync via VPC endpoint
    // Store connection mapping for real-time updates
    
    return { statusCode: 200, body: 'Subscribed' };
}

Step 5: Alternative Simpler Approach - HTTP Streaming
Since WebSocket proxying is complex, consider Server-Sent Events (SSE):

// API Gateway Lambda for SSE
exports.handler = async (event) => {
    const { httpMethod, path, queryStringParameters } = event;
    
    if (httpMethod === 'GET' && path === '/stream') {
        const channelId = queryStringParameters.channelId;
        
        // Return SSE headers
        return {
            statusCode: 200,
            headers: {
                'Content-Type': 'text/event-stream',
                'Cache-Control': 'no-cache',
                'Connection': 'keep-alive',
                'Access-Control-Allow-Origin': '*'
            },
            body: `data: {"type":"connected","channelId":"${channelId}"}\n\n`,
            isBase64Encoded: false
        };
    }
    
    if (httpMethod === 'POST' && path === '/graphql') {
        // Proxy GraphQL mutations to private AppSync
        return await proxyToAppSync(JSON.parse(event.body));
    }
};

async function proxyToAppSync(graphqlRequest) {
    const response = await fetch(PRIVATE_APPSYNC_ENDPOINT, {
        method: 'POST',
        headers: {
            'Content-Type': 'application/json',
            'Authorization': await getIAMSignature()
        },
        body: JSON.stringify(graphqlRequest)
    });
    
    return {
        statusCode: response.status,
        headers: {
            'Content-Type': 'application/json',
            'Access-Control-Allow-Origin': '*'
        },
        body: await response.text()
    };
}

Step 6: Client-Side Implementation
// Client connects through your API Gateway
class AppSyncProxyClient {
    constructor(apiGatewayUrl, channelId) {
        this.apiGatewayUrl = apiGatewayUrl;
        this.channelId = channelId;
        this.eventSource = null;
    }

    // Subscribe using Server-Sent Events
    subscribe(onMessage) {
        const streamUrl = `${this.apiGatewayUrl}/stream?channelId=${this.channelId}`;
        
        this.eventSource = new EventSource(streamUrl);
        
        this.eventSource.onmessage = (event) => {
            const data = JSON.parse(event.data);
            onMessage(data);
        };
        
        this.eventSource.onerror = (error) => {
            console.error('SSE Error:', error);
        };
    }

    // Send mutations through API Gateway
    async sendMutation(mutation, variables) {
        const response = await fetch(`${this.apiGatewayUrl}/graphql`, {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json',
                // No auth headers - handled by Gloo Gateway
            },
            body: JSON.stringify({
                query: mutation,
                variables
            })
        });
        
        return await response.json();
    }

    disconnect() {
        if (this.eventSource) {
            this.eventSource.close();
        }
    }
}

// Usage
const client = new AppSyncProxyClient('https://your-api-gateway-url', 'user-123');

client.subscribe((message) => {
    console.log('Received:', message);
    updateUI(message);
});

Step 7: Lambda in VPC Configuration
// Lambda function that calls private AppSync
const AWS = require('aws-sdk');

// Configure AppSync client to use VPC endpoint
const appsync = new AWS.AppSync({
    endpoint: 'https://vpce-xxxxx-xxxxx.appsync-api.region.vpce.amazonaws.com'
});

exports.handler = async (event) => {
    const mutation = `
        mutation PublishUpdate($input: UpdateInput!) {
            publishUpdate(input: $input) {
                id
                channelId
                message
                timestamp
            }
        }
    `;

    const variables = {
        input: {
            channelId: event.channelId,
            message: event.message,
            sender: 'Bot'
        }
    };

    try {
        // Call private AppSync through VPC endpoint
        const result = await callPrivateAppSync(mutation, variables);
        
        // Also notify API Gateway clients via SSE
        await notifyApiGatewayClients(event.channelId, result);
        
        return { statusCode: 200, body: JSON.stringify(result) };
    } catch (error) {
        console.error('Error:', error);
        return { statusCode: 500, body: JSON.stringify({ error: error.message }) };
    }
};

async function callPrivateAppSync(query, variables) {
    // Use AWS SDK with IAM signing for private AppSync
    const params = {
        apiId: process.env.APPSYNC_API_ID,
        format: 'SDL',
        typeName: 'Mutation',
        fieldName: 'publishUpdate',
        payload: JSON.stringify({ query, variables })
    };
    
    return await appsync.evaluateCode(params).promise();
}

async function notifyApiGatewayClients(channelId, data) {
    // Notify connected API Gateway WebSocket/SSE clients
    // This bridges private AppSync to public API Gateway
}

Step 8: Infrastructure Setup
# CloudFormation template
Resources:
  # Private AppSync API
  PrivateAppSyncApi:
    Type: AWS::AppSync::GraphQLApi
    Properties:
      Name: PrivateChatbotAPI
      AuthenticationType: AWS_IAM
      Visibility: PRIVATE
      
  # VPC Endpoint for AppSync
  AppSyncVPCEndpoint:
    Type: AWS::EC2::VPCEndpoint
    Properties:
      VpcId: !Ref VPC
      ServiceName: !Sub 'com.amazonaws.${AWS::Region}.appsync-api'
      VpcEndpointType: Interface
      SubnetIds:
        - !Ref PrivateSubnet1
        - !Ref PrivateSubnet2
      SecurityGroupIds:
        - !Ref AppSyncVPCEndpointSG

  # API Gateway for proxy
  ProxyApiGateway:
    Type: AWS::ApiGateway::RestApi
    Properties:
      Name: AppSyncProxy
      EndpointConfiguration:
        Types: [REGIONAL]

  # Lambda for API Gateway integration
  ProxyLambda:
    Type: AWS::Lambda::Function
    Properties:
      Runtime: nodejs18.x
      Handler: index.handler
      VpcConfig:
        SubnetIds:
          - !Ref PrivateSubnet1
          - !Ref PrivateSubnet2
        SecurityGroupIds:
          - !Ref LambdaSG

Summary
This architecture provides:

✅ Private AppSync (no public access)
✅ All traffic through USAA network → Gloo Gateway → API Gateway
✅ No authentication at AppSync (handled upstream)
✅ Real-time capabilities via SSE or WebSocket proxy
✅ Lambda in VPC can call private AppSync
✅ Enterprise security compliance
The key is using API Gateway as a proxy to bridge between your public-facing requirements and the private AppSync API, while maintaining real-time capabilities through Server-Sent Events or WebSocket proxying.
