import { BedrockRuntimeClient, InvokeModelCommand } from '@aws-sdk/client-bedrock-runtime';
import { LexRuntimeV2Client, RecognizeTextCommand } from '@aws-sdk/client-lex-runtime-v2';

const bedrock = new BedrockRuntimeClient({});
const lex = new LexRuntimeV2Client({});

const INTENTS = [
  'routing_number',
  'dispute_transaction',
  'credit_rate_interest',
  'pay_bill',
  'transfer_funds',
  'report_card_lost',
  'claim_insurance',
  'coverage_info',
  'branch_hours_location',
  'agent_handoff'
];

const SYSTEM_PROMPT = `You are an intent classifier. Only classify into one of these intents: ${INTENTS.join(', ')}.\nReturn a strict JSON object: {"intent":"<one|unknown>","confidence":0-1,"clarification":"optional question if low confidence"}.\nIf user asks for human, choose agent_handoff.\nIf you are not sure, set intent to "unknown" and provide a short clarifying question.`;

interface ClassificationResult { intent: string; confidence: number; clarification?: string }

async function classify(message: string): Promise<ClassificationResult> {
  if (process.env.USE_BEDROCK !== 'true') {
    return keywordFallback(message);
  }
  const modelId = process.env.BEDROCK_MODEL_ID!;
  try {
    const body = { model: modelId, max_tokens: 150, messages: [ { role: 'system', content: SYSTEM_PROMPT }, { role: 'user', content: message } ], temperature: 0 };
    const cmd = new InvokeModelCommand({ modelId, body: new TextEncoder().encode(JSON.stringify(body)), contentType: 'application/json' });
    const resp = await bedrock.send(cmd);
    const text = new TextDecoder().decode(resp.body as Uint8Array).trim();
    let jsonStr = text;
    const match = jsonStr.match(/\{[\s\S]*\}/);
    if (match) jsonStr = match[0];
    const parsed = JSON.parse(jsonStr);
    return { intent: parsed.intent || 'unknown', confidence: Number(parsed.confidence) || 0, clarification: parsed.clarification };
  } catch (e: any) {
    console.warn('Bedrock classify failed, falling back:', e?.message);
    return keywordFallback(message);
  }
}

function keywordFallback(message: string): ClassificationResult {
  const m = message.toLowerCase();
  const mapping: { intent: string; keywords: string[] }[] = [
    { intent: 'report_card_lost', keywords: ['lost card', 'lost my card', 'block my card', 'stolen card'] },
    { intent: 'dispute_transaction', keywords: ['dispute', 'unauthorized', 'charge i don'] },
    { intent: 'transfer_funds', keywords: ['transfer', 'move money'] },
    { intent: 'pay_bill', keywords: ['pay', 'bill', 'payment'] },
    { intent: 'routing_number', keywords: ['routing number'] },
    { intent: 'claim_insurance', keywords: ['file a claim', 'insurance claim'] },
    { intent: 'coverage_info', keywords: ['what\'s covered', 'coverage'] },
    { intent: 'branch_hours_location', keywords: ['nearest branch', 'branch', 'atm'] },
    { intent: 'agent_handoff', keywords: ['human', 'agent', 'representative'] },
    { intent: 'credit_rate_interest', keywords: ['interest rate', 'apr'] }
  ];
  for (const entry of mapping) {
    if (entry.keywords.some(k => m.includes(k))) {
      return { intent: entry.intent, confidence: 0.9 };
    }
  }
  return { intent: 'unknown', confidence: 0, clarification: 'Could you clarify what you need help with?' };
}

interface ChatEventBody { userId: string; sessionId: string; message: string }

export const handler = async (event: any) => {
  try {
    const body: ChatEventBody = typeof event.body === 'string' ? JSON.parse(event.body) : event.body;
    if (!body?.message || !body?.sessionId) return response(400, { error: 'Missing message or sessionId' });
    const threshold = Number(process.env.INTENT_CONFIDENCE_THRESHOLD || '0.6');
  const classification = await classify(body.message);
    if (classification.intent === 'unknown' || classification.confidence < threshold) {
      return response(200, { type: 'clarification', message: classification.clarification || 'I want to make sure I understand. Can you please provide a little more detail?' });
    }
    const lexResp = await lex.send(new RecognizeTextCommand({
      botId: process.env.LEX_BOT_ID!,
      botAliasId: process.env.LEX_BOT_ALIAS_ID!,
      localeId: process.env.LEX_LOCALE_ID || 'en_US',
      sessionId: body.sessionId,
      text: body.message,
      sessionState: { intent: { name: classification.intent, state: 'InProgress' } }
    }));
    const messages = (lexResp.messages || []).map(m => m?.content || '').filter(Boolean);
    return response(200, { type: 'lex', intent: classification.intent, confidence: classification.confidence, messages });
  } catch (err: any) {
    console.error(err);
    return response(500, { error: 'Internal error', details: err.message });
  }
};

function response(statusCode: number, body: any) {
  return { statusCode, headers: { 'Content-Type': 'application/json', 'Access-Control-Allow-Origin': '*' }, body: JSON.stringify(body) };
}
