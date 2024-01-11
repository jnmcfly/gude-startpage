import { Form, FormControl, Container, Row, Col, Card } from "react-bootstrap";
import { MeiliSearch } from "meilisearch";
import React, { useState, useEffect, useCallback, useRef } from "react";
import { FontAwesomeIcon } from "@fortawesome/react-fontawesome";
import { faTag, faStar } from "@fortawesome/free-solid-svg-icons";

function getGreeting() {
  const currentHour = new Date().getHours();
  let greeting;

  if (currentHour < 12) {
    greeting = "Gude morning!";
  } else if (currentHour < 18) {
    greeting = "Gude!";
  } else {
    greeting = "Gude evening!";
  }

  return greeting;
}

function Search() {
  const [searchTerm, setSearchTerm] = useState("");
  const [searchResults, setSearchResults] = useState([]);
  const [selectedItemIndex, setSelectedItemIndex] = useState(-1);
  const searchResultsRef = useRef(null);

  const handleSearch = useCallback(async () => {
    if (searchTerm === "") {
      setSearchResults([]); // Clear the search results if the search term is empty
      return;
    }

    const client = new MeiliSearch({ host: "http://localhost:7700" });
    const index = client.index("github_stars");
    const results = await index.search(searchTerm, { limit: 7 }); // Limit the results to 7
    setSearchResults(results.hits);
    setSelectedItemIndex(-1); // Reset the selected item index
  }, [searchTerm]);

  useEffect(() => {
    handleSearch();
  }, [handleSearch]);

  const handleKeyDown = (e) => {
    if (e.key === "ArrowUp") {
      e.preventDefault();
      setSelectedItemIndex((prevIndex) =>
        prevIndex > 0 ? prevIndex - 1 : searchResults.length - 1
      );
    } else if (e.key === "ArrowDown") {
      e.preventDefault();
      setSelectedItemIndex((prevIndex) =>
        prevIndex < searchResults.length - 1 ? prevIndex + 1 : 0
      );
    } else if (e.key === "Enter") {
      e.preventDefault();
      if (selectedItemIndex !== -1) {
        const selectedResult = searchResults[selectedItemIndex];
        window.open(selectedResult.url, "_self");
      }
    } else if (e.key === "Escape") {
      e.preventDefault();
      setSearchTerm("");
    }
  };

  return (
    <Container>
      <Container>
        <div className="greeting">{getGreeting()}</div>
      </Container>
      <Row className="justify-content-center">
        <Col md={6}>
          <Form className="text-center search-form">
            <FormControl
              type="text"
              value={searchTerm}
              onChange={(e) => setSearchTerm(e.target.value)}
              onKeyDown={handleKeyDown}
              autoFocus
            />
          </Form>
        </Col>
      </Row>
      {searchTerm !== "" && (
        <Row
          className="justify-content-center mt-4 d-flex"
          style={{ flexWrap: "nowrap" }}
          ref={searchResultsRef}
        >
          {searchResults.map((result, index) => (
            <Col key={result.id} md={4} className={`mb-4`}>
              <Card
                className={`card ${
                  selectedItemIndex === index ? "selected-card" : ""
                }`}
              >
                <Card.Body>
                  <Card.Title className="card-header">
                    <a
                      href={result.url}
                      target="_self"
                      rel="noopener noreferrer"
                      className="search-result-link"
                    >
                      <div className="search-result-link-first">
                        {result.owner}/
                      </div>
                      {result.repository_name}
                    </a>
                  </Card.Title>
                  <div style={{ textAlign: "right", margin: "10px 0" }}>
                    {result.tag_name && (
                      <div className="search-result-detail">
                        <FontAwesomeIcon icon={faTag} /> {result.tag_name}
                      </div>
                    )}
                    {result.stars && (
                      <div className="search-result-stars">
                        <FontAwesomeIcon icon={faStar} /> {result.stars}
                      </div>
                    )}
                  </div>
                </Card.Body>
                {selectedItemIndex === index && (
                  <div className="selected-indicator"></div>
                )}
              </Card>
            </Col>
          ))}
        </Row>
      )}
    </Container>
  );
}

export default Search;
