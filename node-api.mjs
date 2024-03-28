import { spawn } from "child_process";
import path from "path";

const __dirname = new URL(".", import.meta.url).pathname;

/**
 *
 * @param {string} dirPath
 * @param {import("./node-api").ServeOptions} options
 * @param {import("./node-api").ServeSpawnOptions} spawnOptions
 */
export function serve(dirPath, options, spawnOptions) {
  /* @type {string[]} */
  const args = [];

  if (options.port) {
    args.push("--port", String(options.port));
  }
  if (options.loglevel) {
    args.push("--loglevel", options.loglevel);
  }
  if (options.redirect) {
    args.push("--redirect", options.redirect);
  }
  if (options.hmr) {
    if (options.hmr.watch) {
      args.push("--watch");
    }
    if (options.hmr.autoReload) {
      args.push("--auto-reload");
    }
  }
  if (options.cacheHeaders) {
    if (options.cacheHeaders.maxAge) {
      args.push("--maxage", String(options.cacheHeaders.maxAge));
    }
    if (options.cacheHeaders.nocache) {
      args.push("--nocache");
    }
    if (options.cacheHeaders.noEtag) {
      args.push("--noetag");
    }
  }
  if (options.serverCache) {
    if (options.serverCache.max) {
      args.push("--cache:max", String(options.serverCache.max));
    }
    if (options.serverCache.fLimit) {
      args.push("--cache:flimit", String(options.serverCache.fLimit));
    }
  }

  args.push(dirPath);

  return spawn(path.resolve(__dirname, "serve"), args, spawnOptions);
}
