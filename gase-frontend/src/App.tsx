import { Routes, Route } from "react-router-dom";
import { Navbar } from "./components/Navbar";
import { HomePage } from "./pages/HomePage";
import { GasesPage } from "./pages/GasesPage";
import { GasDetailPage } from "./pages/GasDetailPage";
import { ROUTES } from "./Routes";

function App() {
  return (
    <>
      <Navbar />
      <Routes>
        <Route path={ROUTES.HOME} element={<HomePage />} />
        <Route path={ROUTES.GASES} element={<GasesPage />} />
        <Route path={`${ROUTES.GASES}/:id`} element={<GasDetailPage />} />
      </Routes>
    </>
  );
}

export default App;

