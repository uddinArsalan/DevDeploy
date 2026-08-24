import { check, JSONObject } from "k6";
import { Trend } from "k6/metrics";
// k6/x/sse is provided by the k6 SSE extension at runtime and has no TypeScript declarations.
// @ts-ignore
import sse from "k6/x/sse";
import http from "k6/http";
import { User } from "./types.ts";
import { BASE_PRIVATE_URL, BASE_PUBLIC_URL } from "./constants.ts";

export const options = {
  scenarios: {
    default: {
      executor: "per-vu-iterations",
      vus: 10,
      iterations: 5,
    },
  },
};

const deploymentDuration = new Trend("deployment_duration", true);
const totalDuration = new Trend("total_duration", true);

export function setup() {
  const users: User[] = [];
  for (let i = 0; i < 10; i++) {
    const payload = {
      email: `load-user-${i}@example.com`,
      password: `LoadTest@753_${i}`,
    };
    // ----- REGISTER ------
    const regRes = http.post(
      `${BASE_PUBLIC_URL}/auth/signup`,
      JSON.stringify({
        name: `Load-User-${i}`,
        ...payload,
      }),
    );

    const isUserCreated = check(regRes, {
      "User Created": (regRes) => regRes.status == 201,
    });

    if (!isUserCreated) {
      console.error(`Register failed: ${regRes.status} ${regRes.body}`);
      continue;
    }

    users.push(payload);
  }
  return users;
}

export default function (users: User[]) {
  const start = Date.now();
  const user = users[__VU - 1];
  // ----- LOGIN ------
  const url = `${BASE_PUBLIC_URL}/auth/signin`;

  const params = {
    headers: {
      "Content-Type": "application/json",
    },
  };

  const loginRes = http.post(
    url,
    JSON.stringify({
      email: user.email,
      password: user.password,
    }),
    params,
  );

  const isLoggedIn = check(loginRes, {
    "User Logged in": (loginRes) => loginRes.status == 200,
  });

  if (!isLoggedIn) {
    console.error(`Login failed: ${loginRes.status} ${loginRes.body}`);
    return;
  }

  // ------ CREATE PROJECT ------
  const projRes = http.post(
    `${BASE_PRIVATE_URL}/project`,
    JSON.stringify({
      name: `Load-Project-${__VU}-${__ITER}`,
      git_url: "https://github.com/uddinArsalan/ProfilePro",
    }),
    params,
  );

  const projectCreated = check(projRes, {
    "Project created": (r) => r.status === 200,
  });

  if (!projectCreated) {
    console.error(`Project creation failed: ${projRes.status} ${projRes.body}`);
    return;
  }

  const projectRes = projRes.json() as {
    data: {
      project_id: number;
    };
  };

  const projectID = projectRes.data.project_id;

  const deploymentStart = Date.now();

  // Deploy Pipeline
  const deployRes = http.post(
    `${BASE_PRIVATE_URL}/projects/${projectID}/deployments`,
    {},
    params,
  );
  const isDeploymentStarted = check(deployRes, {
    "deployment created": (deployRes) => deployRes.status == 200,
  });
  if (!isDeploymentStarted) {
    console.error(`Deployment failed: ${deployRes.status} ${deployRes.body}`);
    return;
  }
  const deployResp = deployRes.json() as {
    data: {
      deploy_id: number;
    };
  };

  const deployID = deployResp.data.deploy_id;

  const sseUrl = `${BASE_PRIVATE_URL}/stream/${deployID}`;

  const response = sse.open(
    sseUrl,
    {
      method: "GET",
    },
    function (client: any) {
      client.on("open", function open() {
        console.log("connected");
      });

      client.on("event", function (event: any) {
        if (event.type === "done") {
          console.log(`Deployment finished: ${event.data}`);
          client.close();
        }
      });

      client.on("error", function (e: any) {
        console.log("An unexpected error occurred: ", e.error());
      });
    },
  );

  check(response, { "Deployment finished": (r) => r && r.status === 200 });

  const end = Date.now();
  deploymentDuration.add(end - deploymentStart);

  totalDuration.add(end - start);
}
