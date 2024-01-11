import React from "react";
import { Container } from "react-bootstrap";
import Search from "./Search";

const App = () => {
  return (
    <div
      className="App"
      style={{ display: "flex", justifyContent: "center", marginTop: "100px" }}
    >
      <Container style={{ width: "35%" }}>
        <Search />
      </Container>
    </div>
  );
};

export default App;
