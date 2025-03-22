import React from 'react';

const Help: React.FC = () => {
  return (
    <div className="help-page">
      <div className="help-header"></div>
      <div className="help-content">
        <section>
          <h2>Getting Started</h2>
          <p>
            Welcome to game2d.ai! This platform allows you to create, play and share 2D games using
            generative AI.
            Below you'll find answers to common questions and guides to help you get started.
          </p>
        </section>

        <section>
          <h2>Controls and Navigation</h2>
          <p>
            The game2d.ai interface is designed to be intuitive and easy to use. The main navigation is
            located at the top of the screen, allowing you to quickly access different sections of the
            platform.
          </p>
          <p>
            While playing a game, press <span className="keys">Ctrl + S</span> to save the current game
            state to the server. Press <span className="keys">Ctrl + L</span> to load the last saved game
            state. Press <span className="keys">Ctrl + '</span> to display debug information about the
            game. And, finally, press <span className="keys">Ctrl + P</span> to pause the game and press
            <span className="keys">Ctrl + Q</span> to reset the game.
          </p>
          <p>
            For game-specific controls, refer to the individual game's instructions which will be displayed
            in the description of the game.
          </p>
        </section>

        <section>
          <h2>Documentation and Source Code</h2>
          <p>
            The game2d.ai platform is open-source and available on GitHub. You can find the source code and
            documentation at <a href="https://github.com/dhaifley/game2d">this link</a>.
          </p>
          <p>
            We encourage contributions and feedback from the community.
          </p>
          <p>
            The documentation for the game2d.ai API can be accessed <a href="/api/v1/docs">here</a>.
          </p>
        </section>
        
        <section>
          <h2>FAQ</h2>
          <div className="faq-item">
            <h3>How do I create a new game?</h3>
            <p>
              To create a new game, navigate to the Games page after signing in and click the "New" button.
              You'll be guided through the process of setting up your new game definition.
            </p>
          </div>
          
          <div className="faq-item">
            <h3>Can I share my games with others?</h3>
            <p>
              Coming soon! All games you create will be able to be shared with other users. You can control
              individual game visibility using the Public checkbox when you edit your games.
            </p>
          </div>
          
          <div className="faq-item">
            <h3>How do I report issues?</h3>
            <p>
              If you encounter any problems while using the platform, please report them on our GitHub
              issues page at <a href="https://github.com/dhaifley/game2d/issues">this link</a>.
              We appreciate your feedback and will work to resolve any issues as quickly as possible.
            </p>
          </div>
        </section>
      </div>
    </div>
  );
};

export default Help;
