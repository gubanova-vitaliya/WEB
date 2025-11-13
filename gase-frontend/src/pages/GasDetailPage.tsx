import "./GasDetailPage.css";
import { FC, useEffect, useState } from "react";
import { BreadCrumbs } from "../components/Breadcrumbs";
import { ROUTES, ROUTE_LABELS } from "../Routes";
import { useParams, useNavigate } from "react-router-dom";
import { Gas } from "../components/GasCard";
import { getGasById } from "../modules/gasApi";
import { Spinner } from "react-bootstrap";
import defaultImage from "/DefaultImage.svg";

export const GasDetailPage: FC = () => {
  const [pageData, setPageData] = useState<Gas | null>(null);
  const [loading, setLoading] = useState(true);

  const { id } = useParams<{ id: string }>();
  const navigate = useNavigate();

  useEffect(() => {
    if (!id) return;

    setLoading(true);
    getGasById(parseInt(id))
      .then((gas) => {
        setPageData(gas);
      })
      .catch((error) => {
        console.error("Error loading gas:", error);
      })
      .finally(() => {
        setLoading(false);
      });
  }, [id]);

  if (loading) {
    return (
      <div className="gas-detail-page">
        <BreadCrumbs
          crumbs={[
            { label: ROUTE_LABELS.GASES, path: ROUTES.GASES },
            { label: "Загрузка..." },
          ]}
        />
        <div className="loader-block">
          <Spinner animation="border" />
        </div>
      </div>
    );
  }

  if (!pageData) {
    return (
      <div className="gas-detail-page">
        <BreadCrumbs
          crumbs={[
            { label: ROUTE_LABELS.GASES, path: ROUTES.GASES },
            { label: "Газ не найден" },
          ]}
        />
        <div className="gas-detail-container">
          <div className="gas-card">
            <div className="not-found">
              <h2>Газ не найден</h2>
              <p>Запрошенный газ не существует или был удален</p>
              <a href={ROUTES.GASES} className="yellow-btn" onClick={(e) => {
                e.preventDefault();
                navigate(ROUTES.GASES);
              }}>
                Вернуться к каталогу
              </a>
            </div>
          </div>
        </div>
      </div>
    );
  }

  return (
    <div className="gas-detail-page">
      <BreadCrumbs
        crumbs={[
          { label: ROUTE_LABELS.GASES, path: ROUTES.GASES },
          { label: pageData.title },
        ]}
      />
      <div className="gas-detail-container">
        <div className="gas-card">
          <div className="gas-header">
            <h1>{pageData.title}</h1>
            <span className="gas-formula">{pageData.formula}</span>
          </div>

          <div className="gas-image">
            <img
              src={pageData.image_url || defaultImage}
              alt={pageData.title}
              onError={(e) => {
                (e.target as HTMLImageElement).src = defaultImage;
              }}
            />
          </div>

          <div className="gas-properties">
            <div className="property-card">
              <div className="property-label">Молярная масса</div>
              <div className="property-value">
                {pageData.molar_mass.toFixed(2)} г/моль
              </div>
            </div>
            <div className="property-card">
              <div className="property-label">ID газа</div>
              <div className="property-value">#{pageData.id}</div>
            </div>
          </div>

          <div className="gas-description">
            {pageData.description || "Описание отсутствует"}
          </div>
        </div>
      </div>
    </div>
  );
};
