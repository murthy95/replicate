import Footer from "../components/footer";
import Header from "../components/header";
import Layout from "./default";
import Link from "next/link";

function DocsLayout({ title, children, ...props }) {
  return (
    <Layout title={title || "Documentation"} {...props}>
      <Header className="documentation">
        <div className="breadcrumb">
          <Link href="/">
            <a>Home</a>
          </Link>
          &nbsp;
          {title ? (
            <>
              <Link href="/docs">
                <a>
                  <span>Documentation</span>
                </a>
              </Link>
              &nbsp;<h2>{title}</h2>
            </>
          ) : (
            <h2>Documentation</h2>
          )}
        </div>
      </Header>

      <section className="docs documentation">
        <nav>
          <ol>
            <li>
              <ol>
                <li>
                  <Link href="/docs">
                    <a>Install &amp; first steps</a>
                  </Link>
                </li>
                <li>
                  <Link href="/docs/tutorial">
                    <a>Tutorial</a>
                  </Link>
                </li>
                <li>
                  <a href={process.env.TUTORIAL_COLAB_URL} target="_blank">
                    Notebook tutorial
                  </a>
                </li>
              </ol>
            </li>
            <li>
              <h2>Guides</h2>
              <ol>
                <li>
                  <Link href="/docs/guides/cloud-storage">
                    <a>Store data in the cloud</a>
                  </Link>
                </li>
                <li>
                  <Link href="/docs/guides/training-data">
                    <a>Version training data</a>
                  </Link>
                </li>
                <li>
                  <a href={process.env.ANALYSIS_COLAB_URL} target="_blank">
                    Analyze &amp; visualize in a notebook
                  </a>
                </li>
                <li>
                  <Link href="/docs/guides/keras-integration">
                    <a>Keras integration</a>
                  </Link>
                </li>
                <li>
                  <Link href="/docs/guides/pytorch-lightning-integration">
                    <a>PyTorch Lightning integration</a>
                  </Link>
                </li>
                <li>
                  <Link href="/docs/guides/inference">
                    <a>Load models for inference</a>
                  </Link>
                </li>
              </ol>
            </li>
            <li>
              <h2>Learning</h2>
              <ol>
                <li>
                  <Link href="/docs/learn/how-it-works">
                    <a>How it works</a>
                  </Link>
                </li>
                <li>
                  <Link href="/docs/learn/analytics">
                    <a>Analytics</a>
                  </Link>
                </li>
              </ol>
            </li>
            <li>
              <h2>Reference</h2>
              <ol>
                <li>
                  <Link href="/docs/reference/python">
                    <a>Python library</a>
                  </Link>
                </li>
                <li>
                  <Link href="/docs/reference/yaml">
                    <a>keepsake.yaml</a>
                  </Link>
                </li>
                <li>
                  <Link href="/docs/reference/cli">
                    <a>Command-line interface</a>
                  </Link>
                </li>
              </ol>
            </li>
          </ol>
        </nav>
        <div className="body">{children}</div>
      </section>
      <Footer />
    </Layout>
  );
}
export default DocsLayout;
