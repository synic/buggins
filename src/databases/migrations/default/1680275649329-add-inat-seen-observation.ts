import { MigrationInterface, QueryRunner } from 'typeorm';

export class addInatSeenObservation1680275649329 implements MigrationInterface {
  name = 'addInatSeenObservation1680275649329';

  public async up(queryRunner: QueryRunner): Promise<void> {
    await queryRunner.query(
      `CREATE TABLE "inaturalist_seen_observation" ("id" SERIAL NOT NULL, "created_at" TIMESTAMP NOT NULL DEFAULT now(), "updated_at" TIMESTAMP NOT NULL DEFAULT now(), "observation_id" INTEGER NOT NULL, CONSTRAINT "PK_48873114a697b1045bd39c549ca" PRIMARY KEY ("id"))`,
    );
  }

  public async down(queryRunner: QueryRunner): Promise<void> {
    await queryRunner.query(`DROP TABLE "inaturalist_seen_observation"`);
  }
}
