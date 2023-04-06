import { DiscordSettingsBaseEntity } from '@buggins/discord/discord-settings-base.entity';
import { Column, Entity } from 'typeorm';

@Entity({ name: 'inaturalist_settings' })
export class INaturalistSettingsEntity extends DiscordSettingsBaseEntity {
  @Column()
  projectId: string;

  /**
   * iNaturalist Project ID
   */
  @Column()
  channelName: string;

  @Column({ nullable: true })
  cronPattern?: string;
}
