import clsx from "clsx";
import Heading from "@theme/Heading";
import styles from "./styles.module.css";

type FeatureItem = {
  title: string;
  description: JSX.Element;
};

const FeatureList: FeatureItem[] = [
  {
    title: "Multiple protocols support",
    description: <>DNS, HTTP(s), FTP, SMTP</>,
  },
  {
    title: "Custom DNS answers",
    description: (
      <>
        Configure DNS answers with the ability to return multiple records
        for a name or set up DNS rebinding
      </>
    ),
  },
  {
    title: "Custom HTTP responses",
    description: (
      <>
        Configure any HTTP responses: static or dynamic using Go template
        language
      </>
    ),
  },
  {
    title: "Different messengers support",
    description: (
      <>
        Receive notifications to you favourite messenger.
      </>
    ),
  },
  {
    title: "Flexible control",
    description: (
      <>
        Use favourite messenger or CLI to manage your payloads.
      </>
    ),
  },
  {
    title: "REST API",
    description: (
      <>
        Automate vulnerability scanning with the included REST API.
      </>
    ),
  },
];

function Feature({ title, description }: FeatureItem) {
  return (
    <div className={clsx("col col--4")}>
      <div className="text--center padding-horiz--md">
        <Heading as="h3">{title}</Heading>
        <p>{description}</p>
      </div>
    </div>
  );
}

export default function HomepageFeatures(): JSX.Element {
  return (
    <section className={styles.features}>
      <div className="container">
        <div className="row">
          {FeatureList.map((props, idx) => (
            <Feature key={idx} {...props} />
          ))}
        </div>
      </div>
    </section>
  );
}
