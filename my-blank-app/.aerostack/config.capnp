
using Workerd = import "/workerd/workerd.capnp";

const config :Workerd.Config = (
  services = [
    (name = "main", worker = .mainWorker),
  ],
  sockets = [
    ( name = "http",
      address = "*:8787",
      http = (),
      service = "main"
    ),
  ],
);

const mainWorker :Workerd.Worker = (
  modules = [
    (name = "main", esModule = embed "dist/index.js"),
  ],
  compatibilityDate = "2024-01-01",
);
