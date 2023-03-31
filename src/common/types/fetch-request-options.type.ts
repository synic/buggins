export type FetchRequestOptions = {
  // @default - POST
  method?: string;
  body?: string | object;
  path: string;
  // query string parameters
  params?: { [key: string]: string | number | undefined };
};
