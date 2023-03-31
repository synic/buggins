import {
  Client,
  EmbedBuilder,
  Guild,
  Interaction,
  TextBasedChannel,
} from 'discord.js';
import { DiscordModule, Observation } from '@ao/types';
import { Result, Ok } from 'ts-results';
import { FetchCommunicationError } from '@ao/common/types';
import { httpRequest } from '@ao/common/utils';

export class INaturalistModule extends DiscordModule {
  constructor(client: Client) {
    super(client);

    this.addCommand({
      name: 'loadinat',
      description: 'Load inaturalist observations',
      execute: async (interaction: Interaction) =>
        await this.fetch(interaction),
      autoreply: true,
    });
  }

  private async fetchRecentProbjectObservations(): Promise<
    Result<Observation[], FetchCommunicationError>
  > {
    const response = await httpRequest<{ results: Observation[] }>({
      server: 'https://api.inaturalist.org',
      path: `v1/observations?project_id=${
        process.env.PROJECT_ID ?? ''
      }&ttl=900&v=1680225091000&preferred_place_id=52&locale=en&return_bounds=true&per_page=50`,
    });

    if (!response.ok) return response;

    return Ok(response.val.results);
  }

  private async haveSeenObservation(o: Observation): Promise<boolean> {
    if (await this.storage.getItem(`inaturalist-observation-seen-${o.id}`)) {
      return true;
    }

    return false;
  }

  private async markObservationAsSeen(o: Observation): Promise<void> {
    await this.storage.setItem(`inaturalist-observation-seen-${o.id}`, 'true');
  }

  private getChannel(guild: Guild): TextBasedChannel | null {
    return guild.channels.cache.find(
      (c) => c.name === (process.env.INATURALIST_CHANNEL ?? 'inaturalist'),
    ) as TextBasedChannel;
  }

  private async showObservation(
    interaction: Interaction,
    o: Observation,
  ): Promise<void> {
    if (!interaction.guild) throw 'Interaction did not have a guild.';
    const channel = this.getChannel(interaction.guild);
    console.log(channel);
    const photoUrl = o.photos[0].url.replace('square', 'large');
    const image = new EmbedBuilder({ image: { url: photoUrl } });
    await channel?.send({
      content: 'test',
      embeds: [image],
    });

    return;
  }

  async fetch(interaction: Interaction): Promise<void> {
    const observationsResponse = await this.fetchRecentProbjectObservations();
    if (!interaction.channel) throw 'Did not come from a channel';

    if (!observationsResponse.ok) {
      throw observationsResponse.val;
    }

    const shuffled = observationsResponse.val
      .map((value) => ({ value, sort: Math.random() }))
      .sort((a, b) => a.sort - b.sort)
      .map(({ value }) => value);

    while (shuffled.length > 0) {
      const observation = shuffled.pop();

      if (observation?.photos.length) {
        if (!(await this.haveSeenObservation(observation))) {
          await this.showObservation(interaction, observation);
          await this.markObservationAsSeen(observation);
          break;
        } else {
          console.log('Observation already seen');
        }
      }
    }

    return;
  }
}
