import { DiscordModule } from '@ao/types';
import {
  Client,
  MessageReaction,
  Message,
  Events,
  User,
  TextChannel,
} from 'discord.js';

export class GalleryModule extends DiscordModule {
  constructor(client: Client) {
    super(client);

    client.on(Events.MessageCreate, async (message: Message) => {
      const [galleryChannel, _] = this.getChannels(message);

      if (message.channelId === galleryChannel?.id && !message.author.bot) {
        if (message.attachments.size === 1) {
          await message.react('👍');
          await message.react('👎');
          await message.react('💬');
        } else {
          await message.delete();
        }
      }
    });

    client.on(
      Events.MessageReactionAdd,
      async (reaction: MessageReaction, user: User) => {
        const message = reaction.message.partial
          ? await reaction.message.fetch()
          : reaction.message;

        const [galleryChannel, feedbackChannel] = this.getChannels(message);

        const attachment = reaction.message.attachments.first();
        if (
          !user.bot &&
          reaction.emoji.name === '💬' &&
          reaction.message.channelId === galleryChannel?.id &&
          reaction.message.author &&
          attachment
        ) {
          await (feedbackChannel as TextChannel).send({
            content:
              `Hey ${reaction.message.author.toString()}, ${user.toString()} would like to give you ` +
              `feedback on your image.\n Click here for the original: ${reaction.message.url}`,
            files: [attachment],
          });
        }
      },
    );
  }

  getChannels(message: Message): [TextChannel | null, TextChannel | null] {
    const galleryChannel = message.guild?.channels.cache.find(
      (c) => c.name === 'want-feedback',
    );
    const feedbackChannel = message.guild?.channels.cache.find(
      (c) => c.name === 'photo-feedback',
    );
    return [galleryChannel as TextChannel, feedbackChannel as TextChannel];
  }
}
