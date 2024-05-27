const redis = require('redis');
const client = redis.createClient();
const protobuf = require('protobufjs');

// Load the protobuf definition
const root = protobuf.loadSync('message.proto');
const QueueMessage = root.lookupType('QueueMessage');

// Connect to Redis
client.on('error', (err) => console.log('Redis Client Error', err));
client.connect();

// Function to generate a random queue message
function generateQueueMessage() {
  const gateway = ['zain', 'stc', 'mobily'][Math.floor(Math.random() * 3)];
  const phoneNumber = `+966${Math.floor(Math.random() * 1000000000)}`;
  const text = `This is a sample text message from ${gateway}.`;
  return { gateway, phone_number: phoneNumber, text };
}

// Function to write messages to the Redis queue
async function writeToQueue(queueName, messageRate, duration) {
  const queueKey = `queue:${queueName}`;
  const start = new Date().getTime();
  let messageSent = 0;

  while (true) {
    const messageData = generateQueueMessage();
    const message = QueueMessage.create(messageData);
    const buffer = QueueMessage.encode(message).finish();

    await client.rPush(queueKey, buffer);
    messageSent++;

    const now = new Date().getTime();
    const elapsedTime = now - start;

    if (duration && elapsedTime >= duration * 1000) {
      console.log(`Sent ${messageSent} messages to ${queueName} in ${duration} seconds.`);
      break;
    }

    const targetRate = Math.floor(Math.random() * (900 - 100 + 1)) + 100; // Random rate between 100 and 900 messages per second
    const targetDelay = 1000 / targetRate;
    await new Promise((resolve) => setTimeout(resolve, targetDelay));
  }
}

// Main function
async function main() {
  const queueName = 'go-queue-testing-proto';
  const duration = process.argv[2] ? parseInt(process.argv[2]) : null; // Duration in seconds (null for indefinite)

  console.log(`Writing messages to Redis queue '${queueName}'...`);

  if (duration) {
    console.log(`Running for ${duration} seconds.`);
  } else {
    console.log('Running indefinitely. Press Ctrl+C to stop.');
  }

  await writeToQueue(queueName, 100, duration);
  client.quit();
}

main();