import { registerAs } from '@nestjs/config';

export default registerAs('gallery', () => ({
  postChannelName: process.env.GALLERY_POST_CHANNEL_NAME ?? 'want-feedback',
  feedbackChannelName:
    process.env.GALLERY_FEEDBACK_CHANNEL_NAME ?? 'photo-feedback',
}));
