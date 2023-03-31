import { registerAs } from '@nestjs/config';

export default registerAs('discord', () => ({
  token: process.env.BOT_TOKEN ?? '',
}));
