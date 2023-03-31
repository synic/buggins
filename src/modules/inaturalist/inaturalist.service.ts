import { Inject, Injectable, Logger, OnModuleInit } from '@nestjs/common';
import { ConfigType } from '@nestjs/config';
import { Repository } from 'typeorm';
import { InjectRepository } from '@nestjs/typeorm';
import { EmbedBuilder, TextChannel } from 'discord.js';
import { Observation } from './types';
import { Result, Ok } from 'ts-results';
import { FetchCommunicationError } from '@ao/common/types';
import { httpRequest, shuffleArray } from '@ao/common/utils';
import { SeenObservation } from './seen-observation.entity';
import inaturalistConfig from './inaturalist.config';
import { schedule } from 'node-cron';
import { DiscordService } from '@ao/discord/discord.service';

@Injectable()
export class INaturalistService implements OnModuleInit {
  private readonly logger = new Logger(INaturalistService.name);

  constructor(
    private readonly discordService: DiscordService,
    @Inject(inaturalistConfig.KEY)
    private readonly config: ConfigType<typeof inaturalistConfig>,
    @InjectRepository(SeenObservation)
    private readonly seenObservationsRepository: Repository<SeenObservation>,
  ) {}

  onModuleInit() {
    schedule(this.config.cronPattern, () => this.fetch());
    this.logger.log(
      `Set up iNaturalist fetch cronjob with pattern: ${this.config.cronPattern}`,
    );
  }

  private async fetchRecentProjectObservations(): Promise<
    Result<Observation[], FetchCommunicationError>
  > {
    const response = await httpRequest<Observation[]>({
      server: 'https://inaturalist.org',
      path: `observations/project/${this.config.projectId}.json?order_by=id&order=desc`,
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

  private async showObservation(o: Observation): Promise<void> {
    const channel = this.discordService.findChannelByName<TextChannel>(
      this.config.channelName,
    );

    if (!channel) {
      this.logger.error(
        `Could not find channel named: ${this.config.channelName}`,
      );
      return;
    }

    const photoUrl = o.photos[0].large_url;
    const embed = new EmbedBuilder({
      description: `**[${o.user_login}](https://inaturalist.org/people/${o.user_id}) has spotted something new!**`,
    });

    embed.addFields([
      {
        name: `*Species*`,
        value: `${o.species_guess || 'unknown'}`,
      },
      {
        name: 'iNaturalist Link',
        value: `https://inaturalist.org/observations/${o.id}`,
      },
      {
        name: 'iNaturalist Project',
        value: `https://inaturalist.org/projects/${this.config.projectId}`,
      },
    ]);
    embed.setImage(photoUrl);

    await channel?.send({ embeds: [embed] });

    return;
  }

  async fetch(): Promise<void> {
    const observationsResponse = await this.fetchRecentProjectObservations();

    if (!observationsResponse.ok) {
      throw observationsResponse.val;
    }

    const shuffled = shuffleArray<Observation>(observationsResponse.val);

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
