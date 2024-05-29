### Conversational Desigin
- Set Expectation around the constraints of the application
- Anticipate breaking points and accommodate for them.
- Use context and memory  to modify responses.
- Conform to the way that humans speak, and not the other way around.
- Conversational Interface: voice, text, chatbot, other platforms.
- Identify use cases:
     UseCase:   As a __member__ I want to __know the contact number___ so that _I can call with regard to my some questions_.
     ![alt text](image-1.png)

     Problem: member want to know contact address/number
     opportunity:
- Define the persona: brand guidlines. secure and trustworthy financial products, innovatives.
- script the happy path: Open-ended, Menu, form-fillings. how to present? confirming the intent?
![alt text](image.png)

- map flow documenting:
 ![alt text](image-2.png)

- interaction model
![alt text](image-3.png)

- prototyping
  ![alt text](image-4.png)

- test and iterate.




### Amazon Lex for NLU
Amazon Lex provides advanced natural language understanding (NLU) capabilities, which can interpret user input and manage conversation flows.
 - Integration with AWS Services: Lex integrates seamlessly with other AWS services, such as Lambda for executing business logic and AWS Step Functions for workflow orchestrations.
 - Ease of Setup: With Lex, setting up intents, utterances, and slots is relatively straightforward, providing a robust framework for managing various conversational scenarios.
 - Goal oriented dialog system that works by capturing utterances, intents and slots

**Single Bot with Multiple Intents**
Centralized Management: Easier to manage and maintain a single bot with all related intents.
Consistent User Experience: Provides a seamless conversation experience as the user can interact with the bot on various topics without switching.
Efficiency: Reduces overhead and complexity compared to managing multiple bots.
Each intent within the bot is designed to handle specific tasks or responses, and the bot uses these intents to understand and process user inputs effectively.

### Development Deployment strategy
![alt text](image-5.png)


### Rollback Strategy
- in console you can export and import back.
- through terraform as automation.
  -




## Resources:
- [How to approach conversation design: The basics (Part 1)](https://aws.amazon.com/blogs/machine-learning/part-1-approach-conversation-design-the-basics/)
- [How to approach conversation design: Getting started with Amazon Lex (Part 2)](https://aws.amazon.com/blogs/machine-learning/part-2-how-to-approach-conversation-design-getting-started-with-amazon-lex/)
- [Drive efficiencies with CI/CD best practices on Amazon Lex](https://aws.amazon.com/blogs/machine-learning/)drive-efficiencies-with-ci-cd-best-practices-on-amazon-lex/
- https://www.youtube.com/watch?v=0zdU0W22F0s
- [Amazon Lex Model Building Service](https://docs.aws.amazon.com/lex/latest/dg/API_Operations_Amazon_Lex_Model_Building_Service.html)
- [Cost and Pricing](https://aws.amazon.com/lex/pricing/)
- [How it works](https://docs.aws.amazon.com/lexv2/latest/dg/how-it-works.html)
