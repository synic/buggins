import { Column, Entity } from 'typeorm';

import { TimestampEntity } from '@buggins/databases/types';

@Entity({ name: 'inaturalist_seen_observation' })
export class SeenObservationEntity extends TimestampEntity {
  @Column()
  observationId: number;
}
