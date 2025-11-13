import { FC } from "react";
import { Link } from "react-router-dom";
import { ROUTES, ROUTE_LABELS } from "../Routes";
import "./Navbar.css";

export const Navbar: FC = () => {
  return (
    <header>
      <Link to={ROUTES.HOME} className="home-btn">
        🏠 Домой
      </Link>
    </header>
  );
};
