const redis = require('redis');

const client = redis.createClient();
const gateways = ['zain', 'stc', 'mobily']

// Connect to Redis
client.on('error', (err) => console.log('Redis Client Error', err));
client.connect();

// QueueMessage struct
class QueueMessage {
  constructor(messageID, provider, sender, phoneNumbers, text) {
    this.message_id = messageID;
    this.provider = provider;
    this.sender = sender;
    this.phone_numbers = phoneNumbers;
    this.text = text;
  }
}

// Function to generate a random list of phone numbers
function generatePhoneNumbers(maxLength = 10) {
  const numNumbers = Math.floor(Math.random() * maxLength) + 1;
  const numbers = [];
  for (let i = 0; i < numNumbers; i++) {
    num = 966 * 1e9 + Math.floor(Math.random() * 1e9);
    numbers.push(num);
  }
  return numbers;
}

// Function to create a QueueMessage object
function createQueueMessage(gateway) {
  return new QueueMessage(
    Math.random().toString(36).substring(2, 15), // Generate random message ID
    gateway,
    'Your Sender Name', // Replace with your sender name
    generatePhoneNumbers(), // Map numbers to 1 for consistency
    `This is a sample text message from ${gateway}.`
  );
}

// Function to write messages to the Redis queue
async function writeToQueue(messageCount) {
  let totalSent = 0;
  const providerCounts = {};

  for (let i = 0; i < messageCount; i++) {
    const gateway = gateways[Math.floor(Math.random() * gateways.length)];
    gatewayQueue = `go-${gateway}`;
    const message = JSON.stringify(createQueueMessage(gateway));
    await client.rPush(gatewayQueue, message);
    totalSent++;
    providerCounts[gateway] = (providerCounts[gateway] || 0) + 1;
  }

  // Log total numbers and provider breakdown
  console.log(`Total Numbers Sent: ${Object.values(providerCounts).reduce((a, b) => a + b, 0)}`);
  for (const provider in providerCounts) {
    console.log(`  - ${provider}: ${providerCounts[provider]}`);
  }

  client.quit();
}

// Main function
async function main() {
  const messageCount = process.argv[2] ? parseInt(process.argv[2]) : null; // Message count (null for indefinite)

  if (messageCount) {
    console.log(`Sending ${messageCount} messages.`);
  } else {
    console.log('Sending messages indefinitely. Press Ctrl+C to stop.');
  }

  await writeToQueue(messageCount || Infinity); // Use Infinity for indefinite
}

main();
