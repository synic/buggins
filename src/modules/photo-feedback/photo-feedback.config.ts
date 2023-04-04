import { registerAs } from '@nestjs/config';

export default registerAs('feedback', () => ({
  postChannelName: process.env.PHOTO_FEEDBACK_POST_CHANNEL ?? 'want-feedback',
  feedbackChannelName:
    process.env.PHOTO_FEEDBACK_FEEDBACK_CHANNEL ?? 'photo-feedback',
}));
