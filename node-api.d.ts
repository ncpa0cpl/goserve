import {
  ChildProcess,
  ChildProcessByStdio,
  ChildProcessWithoutNullStreams,
  SpawnOptions,
  SpawnOptionsWithStdioTuple,
  SpawnOptionsWithoutStdio,
  StdioNull,
  StdioPipe,
} from "child_process";
import { Readable, Writable } from "stream";

export declare interface ServeOptions {
  port?: number;
  loglevel?: "info" | "debug" | "warn" | "error";
  redirect?: string;
  hmr?: {
    watch?: boolean;
    autoReload?: boolean;
  };
  cacheHeaders?: {
    maxAge?: number;
    nocache?: boolean;
    noEtag?: boolean;
  };
  serverCache?: {
    max?: number;
    fLimit?: number;
  };
}

export declare type ServeSpawnOptions =
  | SpawnOptions
  | SpawnOptionsWithoutStdio
  | SpawnOptionsWithStdioTuple<
      StdioPipe | StdioNull,
      StdioPipe | StdioNull,
      StdioPipe | StdioNull
    >;

export declare function serve(dirPath: string): ChildProcess;
export declare function serve(
  dirPath: string,
  options: ServeOptions
): ChildProcess;
export declare function serve(
  dirPath: string,
  options: ServeOptions | undefined,
  spawnOptions: SpawnOptions
): ChildProcess;
export declare function serve(
  dirPath: string,
  options: ServeOptions | undefined,
  spawnOptions: SpawnOptionsWithoutStdio
): ChildProcessWithoutNullStreams;
export declare function serve(
  dirPath: string,
  options: ServeOptions | undefined,
  spawnOptions: SpawnOptionsWithStdioTuple<
    StdioPipe | StdioNull,
    StdioPipe | StdioNull,
    StdioPipe | StdioNull
  >
): ChildProcessByStdio<Writable | null, Readable | null, Readable | null>;
