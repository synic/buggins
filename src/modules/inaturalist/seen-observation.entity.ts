import { Column, Entity, ManyToOne } from 'typeorm';

import { TimestampEntity } from '@buggins/databases/types';
import { GuildEntity } from '@buggins/discord/guild.entity';

@Entity({ name: 'inaturalist_seen_observation' })
export class SeenObservationEntity extends TimestampEntity {
  @ManyToOne(() => GuildEntity, { nullable: false })
  guildEntity?: GuildEntity;

  @Column({ name: 'guild_entity_id' })
  guildId: string;

  @Column()
  observationId: string;
}
