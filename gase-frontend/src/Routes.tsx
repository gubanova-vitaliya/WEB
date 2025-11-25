export const ROUTES = {
  HOME: "/",
  GASES: "/gases",
};

export type RouteKeyType = keyof typeof ROUTES;

export const ROUTE_LABELS: { [key in RouteKeyType]: string } = {
  HOME: "Главная",
  GASES: "Газы",
};


