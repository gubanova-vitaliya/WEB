export const getCartCount = async (): Promise<number> => {
  try {
    const response = await fetch("/api/cart");
    
    // Если ошибка сервера, возвращаем 0
    if (!response.ok) {
      console.warn(`Cart API returned ${response.status}, using default count 0`);
      return 0;
    }
    
    const data = await response.json();
    return data.count || 0;
  } catch (error) {
    console.warn("Error fetching cart count, using default count 0:", error);
    // Возвращаем 0 при ошибке
    return 0;
  }
};

export const addGasToCart = async (gasId: number): Promise<void> => {
  try {
    const response = await fetch(`/api/gases/${gasId}/add-to-draft`, {
      method: "POST",
    });
    
    if (!response.ok) {
      throw new Error(`Failed to add gas to cart: ${response.status}`);
    }
  } catch (error) {
    console.error("Error adding gas to cart:", error);
    throw error;
  }
};
