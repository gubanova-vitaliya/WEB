import { FC } from "react";
import { Card, Button } from "react-bootstrap";
import "./GasCard.css";
import defaultImage from "/DefaultImage.svg";

export interface Gas {
  id: number;
  title: string;
  formula: string;
  molar_mass: number;
  image_url?: string;
  description?: string;
}

interface GasCardProps {
  gas: Gas;
  onCardClick: (id: number) => void;
}

export const GasCard: FC<GasCardProps> = ({ gas, onCardClick }) => {
  const handleImageError = (e: React.SyntheticEvent<HTMLImageElement, Event>) => {
    const target = e.target as HTMLImageElement;
    if (target.src !== defaultImage) {
      target.src = defaultImage;
    }
  };

  return (
    <Card className="gas-card">
      <Card.Img
        className="card-image"
        variant="top"
        src={gas.image_url || defaultImage}
        alt={gas.title}
        onClick={() => onCardClick(gas.id)}
        onError={handleImageError}
        style={{ cursor: "pointer" }}
      />
      <Card.Body>
        <div className="text-style">
          <Card.Title>{gas.title}</Card.Title>
        </div>
        <div className="text-style">
          <Card.Text>
            <strong>Формула:</strong> {gas.formula}
          </Card.Text>
        </div>
        <div className="text-style">
          <Card.Text>
            <strong>Молярная масса:</strong> {gas.molar_mass.toFixed(2)} г/моль
          </Card.Text>
        </div>
        {gas.description && (
          <div className="text-style description">
            <Card.Text>{gas.description}</Card.Text>
          </div>
        )}
        <Button
          className="card-button"
          variant="primary"
          onClick={() => onCardClick(gas.id)}
        >
          Подробнее
        </Button>
      </Card.Body>
    </Card>
  );
};

