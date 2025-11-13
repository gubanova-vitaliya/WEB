import { Gas } from "../components/GasCard";

export const GASES_MOCK: Gas[] = [
  {
    id: 1,
    title: "Водород",
    formula: "H₂",
    molar_mass: 2.016,
    image_url: "/gas-images/hydrogen.svg",
    description: "Самый легкий химический элемент, бесцветный газ без запаха и вкуса.",
  },
  {
    id: 2,
    title: "Кислород",
    formula: "O₂",
    molar_mass: 32.0,
    image_url: "/gas-images/oxygen.svg",
    description: "Жизненно важный газ, необходимый для дыхания большинства живых организмов.",
  },
  {
    id: 3,
    title: "Азот",
    formula: "N₂",
    molar_mass: 28.014,
    image_url: "/gas-images/nitrogen.svg",
    description: "Инертный газ, составляющий основную часть атмосферы Земли.",
  },
  {
    id: 4,
    title: "Углекислый газ",
    formula: "CO₂",
    molar_mass: 44.01,
    image_url: "/gas-images/carbon-dioxide.svg",
    description: "Газ, образующийся при дыхании и сжигании органических веществ.",
  },
  {
    id: 5,
    title: "Метан",
    formula: "CH₄",
    molar_mass: 16.043,
    image_url: "/gas-images/methane.svg",
    description: "Основной компонент природного газа, простейший углеводород.",
  },
  {
    id: 6,
    title: "Гелий",
    formula: "He",
    molar_mass: 4.003,
    image_url: "/gas-images/helium.svg",
    description: "Инертный газ, второй по легкости элемент, используется в воздушных шарах.",
  },
];
