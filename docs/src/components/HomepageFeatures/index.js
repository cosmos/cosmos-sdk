import React from "react";
import clsx from "clsx";
import styles from "./styles.module.css";
import Link from "@docusaurus/Link";


function Feature({ Svg, to, title, description }) {
  return (
    <div className={clsx("col col--4")}>
      <Link to={to}>
        <div className=" text--center">
          <Svg className={styles.featureSvg} role="img" />
        </div>
      </Link>
      <div className="text--center padding-horiz--md">
        <h3 className="py-5">
          <strong>{title}</strong>
        </h3>
        <p className="pb-10">{description}</p>
      </div>
    </div>
  );
}

export default function HomepageFeatures({list}) {
  return (
    <section className={styles.features}>
      <div className="container pt-9">
        <div className="row">
          {list.map((props, idx) => (
            <Feature key={idx} {...props} />
          ))}
        </div>
      </div>
    </section>
  );
}
