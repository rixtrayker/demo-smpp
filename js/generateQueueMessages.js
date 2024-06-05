const redis = require('redis');
const client = redis.createClient();
const gateways = ['zain', 'stc', 'mobily']

// Connect to Redis
client.on('error', (err) => console.log('Redis Client Error', err));
client.connect();
// Function to generate a random list of phone numbers
function generatePhoneNumbers(maxLength = 10) {
  const numNumbers = Math.floor(Math.random() * maxLength) + 1;
  const numbers = [];
  for (let i = 0; i < numNumbers; i++) {
    numbers.push(966 * 1e9 + Math.floor(Math.random() * 1e9));
  }
  return numbers;
}

// Function to create a QueueMessage object
function createQueueMessage(gateway) {
  return {
    message_id: Math.random().toString(36).substring(2, 15), // Generate random message ID
    provider: gateway,
    sender: 'Your Sender Name', // Replace with your sender name
    phone_numbers: generatePhoneNumbers(), // Map numbers to 1 for consistency
    text: `This is a sample text message from ${gateway}.`,
  };
}

// Function to write messages to the Redis queue
async function writeToQueue(queueName, messageCount) {
  const queueKey = `queue:${queueName}`;
  let totalSent = 0;
  const providerCounts = {};

  for (let i = 0; i < messageCount; i++) {
    const gateway = gateways[Math.floor(Math.random() * 3)];
    gatewayQueue = `queue:${queueName}:${gateway}`;
    const message = JSON.stringify(createQueueMessage(gateway));
    await client.rPush(gatewayQueue, message);
    totalSent++;
    providerCounts[gateway] = (providerCounts[gateway] || 0) + 1;
  }

  console.log(`Sent ${totalSent} messages to ${queueName}.`);

  // Log total numbers and provider breakdown
  console.log(`Total Numbers Sent: ${Object.values(providerCounts).reduce((a, b) => a + b, 0)}`);
  for (const provider in providerCounts) {
    console.log(`  - ${provider}: ${providerCounts[provider]}`);
  }

  client.quit();
}

// Main function
async function main() {
  const queueName = 'go-testing';
  const messageCount = process.argv[2] ? parseInt(process.argv[2]) : null; // Message count (null for indefinite)

  console.log(`Writing messages to Redis queue '${queueName}'...`);

  if (messageCount) {
    console.log(`Sending ${messageCount} messages.`);
  } else {
    console.log('Sending messages indefinitely. Press Ctrl+C to stop.');
  }

  await writeToQueue(queueName, messageCount || Infinity); // Use Infinity for indefinite
}

main();
