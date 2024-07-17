export type Observation = {
  id: number;
  user_id: number;
  user_login: string;
  photos: {
    large_url: string;
  }[];
  species_guess?: string;
  taxon?: {
    name?: string;
    common_name?: {
      name?: string;
    };
    default_name?: {
      name?: string;
    };
  };
};
