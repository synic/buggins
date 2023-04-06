import { TimestampEntity } from '@buggins/databases/types';
import { Column, OneToOne } from 'typeorm';
import { GuildEntity } from './guild.entity';

export abstract class DiscordSettingsBaseEntity extends TimestampEntity {
  @OneToOne(() => GuildEntity, { nullable: false, eager: true })
  guildEntity?: GuildEntity;

  @Column({ name: 'guild_entity_id' })
  guildId: string;

  @Column({ default: false, nullable: false })
  isEnabled: boolean;
}
