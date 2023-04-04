#!/usr/bin/env node

const AWS = require("aws-sdk");
const CP = require("child_process");

const client = new AWS.SecretsManager();

const command = process.argv.slice(2)[0];
const args = process.argv.slice(3);

const runCommandWithEnv = (command, args, env={}) => {
  const child = CP.spawn(command, args, { env: { ...process.env, ...env } });

  child.stdout.on("data", (data) => {
    console.log(data.toString());
  });
  child.stderr.on("data", (data) => {
    console.error(data.toString());
  });
}

if(process.env.AWS_SECRETS_BUNDLE_ID) {
  console.log('Secrets bundle ID found, attempting to expand...');
  client.getSecretValue(
    { SecretId: process.env.AWS_SECRETS_BUNDLE_ID },
    (error, data) => {
      if (error) {
        console.error(error);
        process.exit(1);
      }

      const env = JSON.parse(data.SecretString);
      runCommandWithEnv(command, args, env);
    }
  );
}
else {
  console.log('No secrets id bundle found.');
  runCommandWithEnv(command, args);
}
