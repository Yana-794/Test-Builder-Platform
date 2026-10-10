import { Footer } from "../../widgets/landing/footer/Footer";
import { Header } from "../../widgets/landing/header/Header";
import { Features } from "../../widgets/landing/features/Features";
import { FinalCtaSection } from "../../widgets/landing/final-cta/FinalCta";
import { Hero } from "../../widgets/landing/hero/Hero";
import { HowItWorks } from "../../widgets/landing/how-it-works/HowItWorks";
import { TestPreviewSection } from "../../widgets/landing/test-preview/TestPreview";

export const LandingPage = () => {
  return (
    <>
      <Header />
      <main>
        <Hero />
        <Features />
        <HowItWorks/>
        <TestPreviewSection/>
        <FinalCtaSection/>
      </main> 
      <Footer />
    </>
  );
};
