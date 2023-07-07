import { registerAs } from '@nestjs/config';

export default registerAs('thisthat', () => ({
  channelName: process.env.THISTHAT_CHANNEL_NAME ?? 'this-that',
}));
