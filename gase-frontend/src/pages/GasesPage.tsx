import "./GasesPage.css";
import { FC, useState, useEffect } from "react";
import { Spinner } from "react-bootstrap";
import { ROUTES, ROUTE_LABELS } from "../Routes";
import { useNavigate } from "react-router-dom";
import { getGases, GasFilters } from "../modules/gasApi";
import defaultImage from "/DefaultImage.svg";

export const GasesPage: FC = () => {
  const [gases, setGases] = useState<any[]>([]);
  const [loading, setLoading] = useState(false);
  const [searchValue, setSearchValue] = useState("");
  const [cartCount] = useState(0); // По умолчанию 0, не загружаем из API

  const navigate = useNavigate();

  const loadGases = async () => {
    setLoading(true);
    try {
      const filters: GasFilters = searchValue ? { search: searchValue } : {};
      const filteredGases = await getGases(filters);
      setGases(filteredGases);
    } catch (error) {
      console.error("Error loading gases:", error);
      // Уже обработано в getGases, просто логируем
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    loadGases();
    // Счетчик корзины по умолчанию 0, не загружаем из API
  }, []);

  const handleSearch = (e: React.FormEvent) => {
    e.preventDefault();
    loadGases();
  };

  const handleCardClick = (id: number) => {
    navigate(`${ROUTES.GASES}/${id}`);
  };

  const handleImageError = (e: React.SyntheticEvent<HTMLImageElement, Event>) => {
    // Если изображение не загрузилось (из MinIO или другого источника), используем дефолтное
    const target = e.target as HTMLImageElement;
    if (target.src !== defaultImage) {
      target.src = defaultImage;
    }
  };

  return (
    <div className="gases-page">
      <div className="page-header">
          <h1>{ROUTE_LABELS.GASES}</h1>
        <a
          href="#"
          className={`cart-link ${cartCount > 0 ? "active" : "inactive"}`}
          onClick={(e) => {
            e.preventDefault();
            if (cartCount > 0) {
              // Можно добавить переход на страницу журнала
              alert(`В корзине ${cartCount} газов`);
            }
          }}
        >
          <span className="cart-icon">📋</span>
          Журнал расчетов
          <span className="cart-count">{cartCount}</span>
        </a>
      </div>

      <form className="search-form" onSubmit={handleSearch}>
        <input
          type="text"
          placeholder="Поиск газа..."
          value={searchValue}
          onChange={(e) => setSearchValue(e.target.value)}
        />
        <button type="submit">🔍 Поиск газа</button>
      </form>

      {loading && (
        <div className="loading-bg">
          <Spinner animation="border" />
        </div>
      )}

      {!loading && gases.length === 0 && (
        <div className="no-results">
          <h3>Газы не найдены</h3>
          <p>Попробуйте изменить параметры поиска</p>
        </div>
      )}

      {!loading && gases.length > 0 && (
        <div className="grid">
          {gases.map((gas) => (
            <div key={gas.id} className="card">
              <h2>{gas.title}</h2>
              <img
                src={gas.image_url || defaultImage}
                alt={gas.title}
                width="150"
                onError={handleImageError}
                loading="lazy"
              />
              <p>
                <strong>Формула:</strong> {gas.formula}
              </p>
              <p>
                <strong>Молярная масса:</strong> {gas.molar_mass.toFixed(2)} г/моль
              </p>
              {gas.description && (
                <p>
                  <strong>Описание:</strong> {gas.description}
                </p>
              )}
              <p>
                <strong>ID:</strong> {gas.id}
              </p>
              <a
                href={`${ROUTES.GASES}/${gas.id}`}
                className="yellow-btn"
                onClick={(e) => {
                  e.preventDefault();
                  handleCardClick(gas.id);
                }}
              >
                📖 Подробнее
              </a>
            </div>
          ))}
        </div>
      )}
    </div>
  );
};
