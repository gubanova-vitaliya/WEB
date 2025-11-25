import { Gas } from "../components/GasCard";
import { GASES_MOCK } from "./mock";

export interface GasFilters {
  search?: string;
}

export const getGases = async (filters?: GasFilters): Promise<Gas[]> => {
  try {
    const params = new URLSearchParams();
    if (filters?.search) {
      params.append("search", filters.search);
    }

    const queryString = params.toString();
    const url = `/api/gases${queryString ? `?${queryString}` : ""}`;

    const response = await fetch(url);
    
    // Если ошибка сервера, используем mock данные
    if (!response.ok) {
      console.warn(`API returned ${response.status}, using mock data`);
      throw new Error(`API error: ${response.status}`);
    }

    const data = await response.json();
    
    // Проверяем, что получили массив
    if (!Array.isArray(data)) {
      throw new Error("Invalid response format");
    }
    
    return data.map((gas: any) => ({
      id: gas.ID || gas.id,
      title: gas.Title || gas.title,
      formula: gas.Formula || gas.formula,
      molar_mass: gas.MolarMass || gas.molar_mass,
      image_url: gas.ImageURL || gas.image_url,
      description: gas.Description || gas.description,
    }));
  } catch (error: any) {
    // Перехватываем все ошибки: сетевые, 500 и т.д.
    console.warn("Error fetching gases, using mock data:", error.message || error);
    
    // Используем mock данные при любой ошибке
    let mockGases = [...GASES_MOCK];
    
    // Применяем фильтры к mock данным
    if (filters?.search) {
      const searchLower = filters.search.toLowerCase();
      mockGases = mockGases.filter(
        (gas) =>
          gas.title.toLowerCase().includes(searchLower) ||
          gas.formula.toLowerCase().includes(searchLower)
      );
    }

    return mockGases;
  }
};

export const getGasById = async (id: number): Promise<Gas | null> => {
  try {
    const response = await fetch(`/api/gases/${id}`);
    
    // Если ошибка сервера, используем mock данные
    if (!response.ok) {
      console.warn(`API returned ${response.status}, using mock data`);
      throw new Error(`API error: ${response.status}`);
    }

    const data = await response.json();
    return {
      id: data.ID || data.id,
      title: data.Title || data.title,
      formula: data.Formula || data.formula,
      molar_mass: data.MolarMass || data.molar_mass,
      image_url: data.ImageURL || data.image_url,
      description: data.Description || data.description,
    };
  } catch (error: any) {
    console.warn("Error fetching gas, using mock data:", error.message || error);
    // Используем mock данные при ошибке
    const mockGas = GASES_MOCK.find((gas) => gas.id === id);
    return mockGas || null;
  }
};
