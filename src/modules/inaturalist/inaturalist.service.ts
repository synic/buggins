import { Inject, Injectable, Logger } from '@nestjs/common';
import { ConfigType } from '@nestjs/config';
import { In, Repository } from 'typeorm';
import { InjectRepository } from '@nestjs/typeorm';
import { EmbedBuilder, TextChannel } from 'discord.js';
import { Observation } from './types';
import { Result, Ok } from 'ts-results';
import { FetchCommunicationError } from '@buggins/common/types';
import { httpRequest, shuffleArray } from '@buggins/common/utils';
import { SeenObservation } from './seen-observation.entity';
import inaturalistConfig from './inaturalist.config';
import { DiscordService } from '@buggins/discord/discord.service';

@Injectable()
export class INaturalistService {
  private readonly logger = new Logger(INaturalistService.name);

  constructor(
    private readonly discordService: DiscordService,
    @Inject(inaturalistConfig.KEY)
    private readonly config: ConfigType<typeof inaturalistConfig>,
    @InjectRepository(SeenObservation)
    private readonly seenObservationsRepository: Repository<SeenObservation>,
  ) {}

  private async fetchRecentProjectObservations(): Promise<
    Result<Observation[], FetchCommunicationError>
  > {
    const response = await httpRequest<Observation[]>({
      server: 'https://inaturalist.org',
      path: `observations/project/${this.config.projectId}.json?order_by=id&order=desc&per_page=50`,
    });

    if (!response.ok) return response;

    return Ok(response.val);
  }

  private async markObservationAsSeen(o: Observation): Promise<void> {
    await this.seenObservationsRepository.save({ observationId: o.id });
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
        name: `Species`,
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
    await this.markObservationAsSeen(o);

    return;
  }

  async getRandomObservation(
    observations: Observation[],
  ): Promise<Observation | null> {
    const observationIds = observations.map((o) => o.id);
    const seenObservationIds = (
      await this.seenObservationsRepository.findBy({
        observationId: In(observationIds),
      })
    ).map((o) => o.observationId);

    const unseen = observations.filter(
      (o) => !seenObservationIds.includes(o.id),
    );

    this.logger.log(`Seen observation count is ${seenObservationIds.length}`);
    this.logger.log(`Unseen observation count is ${unseen.length}`);

    if (unseen.length <= 0) {
      this.logger.log(`No unseen observations left to post.`);
      return null;
    }

    const userObservationMap = new Map<number, Observation[]>();

    for (const observation of unseen) {
      let userObservations: Observation[] | null =
        userObservationMap.get(observation.user_id) ?? null;
      if (userObservations == null) {
        userObservations = [];
        userObservationMap.set(observation.user_id, userObservations);
      }

      userObservations.push(observation);
    }

    const user = shuffleArray(Array.from(userObservationMap.keys()))[0];
    const userArray = userObservationMap.get(user);

    if (userArray == null) {
      this.logger.warn(`User array was null`);
      return null;
    }

    this.logger.log(`User is ${user}, items for user is ${userArray.length}`);

    return shuffleArray(userArray)[0];
  }

  async fetch(): Promise<void> {
    const observationsResponse = await this.fetchRecentProjectObservations();

    if (!observationsResponse.ok) {
      this.logger.error(
        `Error fetching observations: ${observationsResponse.val}`,
      );
      return;
    }

    const observation = await this.getRandomObservation(
      observationsResponse.val,
    );

    if (!observation) {
      this.logger.log(`No unseen observations to display at this time.`);
      return;
    }
    await this.showObservation(observation);
  }
}
