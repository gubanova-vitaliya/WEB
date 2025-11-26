import { FC } from "react";
import "./HomePage.css";

const introVideoSrc =
  "https://cdn.pixabay.com/video/2020/08/05/46818-447229696_large.mp4";

export const HomePage: FC = () => {
  return (
    <div className="home-page">
      <section className="hero">
        <div className="hero-video">
          <video
            className="intro-video"
            src={introVideoSrc}
            playsInline
            autoPlay
            loop
            muted
            poster="/slide1.svg"
          />
          <div className="video-overlay" />
        </div>
      </section>
    </div>
  );
};
