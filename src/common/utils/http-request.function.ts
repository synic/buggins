
import { Result, Err, Ok } from 'ts-results';
import fetch from 'node-fetch';
import { FetchCommunicationError, FetchRequestOptions } from '../types';


export const httpRequest = async <T>(
  options: FetchRequestOptions & {
    server: string;
    headers?: { [key: string]: string };
  },
): Promise<Result<T, FetchCommunicationError>> => {
  let body: string | undefined;

  if (options.body != null) {
    if (typeof options.body === 'object') body = JSON.stringify(options.body);
    else body = options.body;
  }

  const response = await fetch(`${options.server}/${options.path}`, {
    method: options?.method ?? (body === undefined ? 'GET' : 'POST'),
    headers: options.headers,
    body,
  });

  if (!response.ok) {
    const data = await response.json();
    const msg = {
      statusCode: response.status,
      reason: response.statusText,
      data,
    };
    console.warn('Invalid response received from http request', {
      server: options.server,
      path: options.path,
      ...msg,
    });

    return Err(msg);
  }

  return Ok((await response.json()) as T);
};
