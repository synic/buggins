import { Column, Entity } from 'typeorm';

import { TimestampEntity } from '@ao/databases/types';

@Entity({ name: 'inaturalist_seen_observation' })
export class SeenObservation extends TimestampEntity {
  @Column()
  observationId: string;
}
