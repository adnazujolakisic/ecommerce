export type ArchitectureNode = {
  id: string;
  label: string;
  stack?: string;
  description: string;
  group:
    | "entry"
    | "infra"
    | "service"
    | "data"
    | "queue"
    | "mirrord";
  repoPath?: string;
  zone?: "cluster" | "external" | "local";
};

export type ArchitectureEdge = {
  id: string;
  source: string;
  target: string;
  label?: string;
  intent?: "request" | "data" | "mirrored" | "control";
};

export type ArchitectureZone = {
  id: string;
  label: string;
  description: string;
  nodes: string[];
  border: string;
  background: string;
  accent: string;
};

export const architectureZones: ArchitectureZone[] = [
  {
    id: "local",
    label: "Local Machine",
    description: "Developer laptop running the binary with mirrord-layer inserted.",
    nodes: ["local-process", "mirrord-layer"],
    border: "#60A5FA",
    background: "rgba(191, 219, 254, 0.4)",
    accent: "#3B82F6",
  },
  {
    id: "cluster",
    label: "GKE Cluster",
    description: "Ingress, services, data stores, and mirrord operator running in-cluster.",
    nodes: [
      "ingress",
      "frontend",
      "catalogue",
      "inventory",
      "checkout",
      "order",
      "order-processor",
      "kafka",
      "postgres-catalogue",
      "postgres-inventory",
      "postgres-orders",
      "mirrord-operator",
      "mirrord-agent",
    ],
    border: "#4F46E5",
    background: "rgba(233, 228, 255, 0.4)",
    accent: "#4F46E5",
  },
];

export const architectureNodes: ArchitectureNode[] = [
  {
    id: "user",
    label: "External user",
    stack: "Browser / curl",
    description: "Initiates traffic through the storefront UI.",
    group: "entry",
    zone: "external",
  },
  {
    id: "ingress",
    label: "Ingress + Service",
    stack: "GKE",
    description: "Public entrypoint routing traffic to the Metal Mart frontend.",
    group: "infra",
    zone: "cluster",
  },
  {
    id: "mirrord-operator",
    label: "mirrord Operator",
    stack: "Kubernetes controller",
    description: "Injects the mirrord agent when a developer session is attached.",
    group: "mirrord",
    zone: "cluster",
  },
  {
    id: "mirrord-agent",
    label: "mirrord Agent",
    stack: "Injected sidecar",
    description: "Appears in pods when mirrord sessions run.",
    group: "mirrord",
    zone: "cluster",
  },
  {
    id: "mirrord-layer",
    label: "mirrord-layer",
    stack: "LD_PRELOAD",
    description: "Intercepts libc calls from the local process.",
    group: "mirrord",
    zone: "local",
  },
  {
    id: "local-process",
    label: "Local process",
    stack: "Developer machine",
    description: "Runs your binary with mirrord-layer injected.",
    group: "mirrord",
    zone: "local",
  },
  {
    id: "frontend",
    label: "frontend",
    stack: "React / Vite",
    description: "E-commerce storefront with product catalog, cart, and checkout.",
    group: "service",
    repoPath: "frontend/",
    zone: "cluster",
  },
  {
    id: "catalogue",
    label: "catalogue",
    stack: "Go",
    description: "Product catalog service.",
    group: "service",
    repoPath: "services/catalogue/",
    zone: "cluster",
  },
  {
    id: "inventory",
    label: "inventory",
    stack: "Go",
    description: "Product catalog and stock management.",
    group: "service",
    repoPath: "services/inventory/",
    zone: "cluster",
  },
  {
    id: "checkout",
    label: "checkout",
    stack: "Go",
    description: "Checkout flow: reserve inventory, create order.",
    group: "service",
    repoPath: "services/checkout/",
    zone: "cluster",
  },
  {
    id: "order",
    label: "order",
    stack: "Go",
    description: "Order creation and status API; publishes to Kafka.",
    group: "service",
    repoPath: "services/order/",
    zone: "cluster",
  },
  {
    id: "order-processor",
    label: "order-processor",
    stack: "Go",
    description: "Kafka consumer that updates order status.",
    group: "service",
    repoPath: "services/order-processor/",
    zone: "cluster",
  },
  {
    id: "kafka",
    label: "Kafka",
    stack: "order.created",
    description: "Receives order events from the order service.",
    group: "queue",
    zone: "cluster",
  },
  {
    id: "postgres-catalogue",
    label: "PostgreSQL (catalogue)",
    stack: "Shared server",
    description: "Product catalog database.",
    group: "data",
    zone: "cluster",
  },
  {
    id: "postgres-inventory",
    label: "PostgreSQL (inventory)",
    stack: "Shared server",
    description: "Stock management database.",
    group: "data",
    zone: "cluster",
  },
  {
    id: "postgres-orders",
    label: "PostgreSQL (orders)",
    stack: "Shared server",
    description: "Orders and status database.",
    group: "data",
    zone: "cluster",
  },
];

export const architectureEdges: ArchitectureEdge[] = [
  {
    id: "user-to-ingress",
    source: "user",
    target: "ingress",
    label: "Browse store",
    intent: "request",
  },
  {
    id: "ingress-to-frontend",
    source: "ingress",
    target: "frontend",
    label: "Route to frontend",
    intent: "request",
  },
  {
    id: "frontend-to-catalogue",
    source: "frontend",
    target: "catalogue",
    label: "GET /products",
    intent: "request",
  },
  {
    id: "frontend-to-inventory",
    source: "frontend",
    target: "inventory",
    label: "GET /inventory",
    intent: "request",
  },
  {
    id: "frontend-to-checkout",
    source: "frontend",
    target: "checkout",
    label: "POST /checkout",
    intent: "request",
  },
  {
    id: "checkout-to-inventory",
    source: "checkout",
    target: "inventory",
    label: "Reserve / Confirm",
    intent: "request",
  },
  {
    id: "checkout-to-order",
    source: "checkout",
    target: "order",
    label: "Create order",
    intent: "request",
  },
  {
    id: "order-to-kafka",
    source: "order",
    target: "kafka",
    label: "Publish order.created",
    intent: "data",
  },
  {
    id: "order-to-postgres-orders",
    source: "order",
    target: "postgres-orders",
    label: "Store order",
    intent: "data",
  },
  {
    id: "kafka-to-order-processor",
    source: "kafka",
    target: "order-processor",
    label: "Consume orders",
    intent: "data",
  },
  {
    id: "order-processor-to-order",
    source: "order-processor",
    target: "order",
    label: "PUT /status",
    intent: "request",
  },
  {
    id: "catalogue-to-postgres-catalogue",
    source: "catalogue",
    target: "postgres-catalogue",
    label: "Products",
    intent: "data",
  },
  {
    id: "inventory-to-postgres-inventory",
    source: "inventory",
    target: "postgres-inventory",
    label: "Stock",
    intent: "data",
  },
  {
    id: "layer-to-agent",
    source: "mirrord-layer",
    target: "mirrord-operator",
    intent: "mirrored",
  },
  {
    id: "operator-to-agent-mirrored",
    source: "mirrord-operator",
    target: "mirrord-agent",
    label: "Launch agent",
    intent: "mirrored",
  },
  {
    id: "local-to-layer",
    source: "local-process",
    target: "mirrord-layer",
    label: "LD_PRELOAD hook",
    intent: "mirrored",
  },
  {
    id: "agent-to-target",
    source: "mirrord-agent",
    target: "order",
    label: "Impersonate target pod",
    intent: "mirrored",
  },
];

export const groupPalette: Record<
  ArchitectureNode["group"],
  { background: string; border: string; text: string }
> = {
  entry: { background: "#FFFFFF", border: "#0F172A", text: "#111827" },
  infra: { background: "#FFFFFF", border: "#6B7280", text: "#111827" },
  service: { background: "#FBF8F2", border: "#EA580C", text: "#111827" },
  data: { background: "#FFFFFF", border: "#DC2626", text: "#111827" },
  queue: { background: "#FFFFFF", border: "#CA8A04", text: "#111827" },
  mirrord: { background: "#EEF2FF", border: "#4F46E5", text: "#111827" },
};
