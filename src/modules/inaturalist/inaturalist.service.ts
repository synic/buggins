import { Inject, Injectable, Logger, OnModuleInit } from '@nestjs/common';
import { schedule } from 'node-cron';
import { ConfigType } from '@nestjs/config';
import { Repository } from 'typeorm';
import { InjectRepository } from '@nestjs/typeorm';
import {
  Client,
  EmbedBuilder,
  Guild,
  Interaction,
  TextBasedChannel,
} from 'discord.js';
import { Observation } from './types';
import { Result, Ok } from 'ts-results';
import { FetchCommunicationError } from '@ao/common/types';
import { httpRequest } from '@ao/common/utils';
import { BaseDiscordService } from '@ao/discord/types';
import { DISCORD_CLIENT_PROVIDER } from '@ao/discord/constants';
import { SeenObservation } from './seen-observation.entity';
import inaturalistConfig from './inaturalist.config';

@Injectable()
export class INaturalistService
  extends BaseDiscordService
  implements OnModuleInit
{
  private readonly logger = new Logger(INaturalistService.name);
  private guild: Guild | undefined;

  constructor(
    @Inject(DISCORD_CLIENT_PROVIDER) client: Client,
    @Inject(inaturalistConfig.KEY)
    private readonly config: ConfigType<typeof inaturalistConfig>,
    @InjectRepository(SeenObservation)
    private readonly seenObservationsRepository: Repository<SeenObservation>,
  ) {
    super(client);

    this.addCommand({
      name: 'loadinat',
      description: 'Load inaturalist observations',
      execute: async (interaction: Interaction) =>
        await this.fetch(interaction),
      autoreply: true,
    });
  }

  onModuleInit() {
    this.guild = this.client.guilds.cache.find(
    schedule('0 * * * *', async () => await this.fetch());
  }

  private async fetchRecentProbjectObservations(): Promise<
    Result<Observation[], FetchCommunicationError>
  > {
    const response = await httpRequest<Observation[]>({
      server: 'https://inaturalist.org',
      path: `observations/project/${this.config.projectId}.json`,
    });

    if (!response.ok) return response;

    return Ok(response.val);
  }

  private async haveSeenObservation(o: Observation): Promise<boolean> {
    const seenObservation = await this.seenObservationsRepository.findOneBy({
      observationId: o.id.toString(),
    });
    return seenObservation != null;
  }

  private async markObservationAsSeen(o: Observation): Promise<void> {
    await this.seenObservationsRepository.save({
      observationId: o.id.toString(),
    });
  }

  private getChannel(guild: Guild): TextBasedChannel | null {
    return guild.channels.cache.find(
      (c) => c.name === this.config.channelName,
    ) as TextBasedChannel;
  }

  private async showObservation(o: Observation): Promise<void> {
    const channel = this.getChannel(interaction.guild);
    const photoUrl = o.photos[0].large_url;
    const image = new EmbedBuilder({ image: { url: photoUrl } });
    await channel?.send({
      content: `${o.user.login} has spotted ${
        o.species_guess ?? 'something new'
      }!: https://inaturalist.org/observations/${o.id}`,
      embeds: [image],
    });

    return;
  }

  async fetch(): Promise<void> {
    const observationsResponse = await this.fetchRecentProbjectObservations();

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
          await this.showObservation(observation);
          await this.markObservationAsSeen(observation);
          break;
        } else {
          this.logger.log(`Observation already seen: ${observation.id}`);
        }
      }
    }

    return;
  }
}
