import { registerAs } from '@nestjs/config';
export default registerAs('inaturalist', () => ({
  channelName: process.env.INATURALIST_CHANNEL_NAME ?? 'inaturalist',
  projectId: process.env.INATURALIST_PROJECT_ID ?? '',
  cronPattern: process.env.INATURALIST_CRON_PATTERN ?? '0 * * * *', // once an hour
}));
