import { FC, useState, useEffect } from "react";
import { Link } from "react-router-dom";
import { BreadCrumbs } from "../components/Breadcrumbs";
import { ROUTES, ROUTE_LABELS } from "../Routes";
import "./HomePage.css";

// Изображения для слайд-шоу
import slide1 from "/slide1.svg";
import slide2 from "/slide2.svg";
import slide3 from "/slide3.svg";

const slides = [
  { id: 1, image: slide1, alt: "Газ 1" },
  { id: 2, image: slide2, alt: "Газ 2" },
  { id: 3, image: slide3, alt: "Газ 3" },
];

export const HomePage: FC = () => {
  const [currentSlide, setCurrentSlide] = useState(0);

  useEffect(() => {
    const timer = setInterval(() => {
      setCurrentSlide((prev) => (prev + 1) % slides.length);
    }, 3000); // Меняем слайд каждые 3 секунды

    return () => clearInterval(timer);
  }, []);

  return (
    <div className="home-page">
      <BreadCrumbs crumbs={[]} />
      <div className="slideshow-container">
        <div className="slideshow-wrapper">
          {slides.map((slide, index) => (
            <div
              key={slide.id}
              className={`slide ${index === currentSlide ? "active" : ""}`}
            >
              <img
                src={slide.image}
                alt={slide.alt}
                onError={(e) => {
                  // Fallback если изображение не загрузилось
                  (e.target as HTMLImageElement).src = "/DefaultImage.svg";
                }}
              />
            </div>
          ))}
        </div>
        <div className="slideshow-title">
          <h1>GasProject</h1>
        </div>
        <div className="slideshow-navigation">
          <Link to={ROUTES.GASES} className="nav-button">
            {ROUTE_LABELS.GASES}
          </Link>
        </div>
        <div className="slideshow-indicators">
          {slides.map((_, index) => (
            <button
              key={index}
              className={`indicator ${index === currentSlide ? "active" : ""}`}
              onClick={() => setCurrentSlide(index)}
            />
          ))}
        </div>
      </div>
    </div>
  );
};
