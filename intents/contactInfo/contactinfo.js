
exports.handler = async (event) => {
  console.dir(event.sessionState);
    console.log("world slots", event.sessionState.intent.slots);

  const accountType = event.sessionState.intent.slots.accType.value.interpretedValue.toLowerCase();
  let accountDetails = '';

  // Dummy data for example purposes
  const accounts = {
    "checking": ["Checking Account 1: $1,200", "Checking Account 2: $2,500"],
    "savings": ["Savings Account 1: $5,000", "Savings Account 2: $10,000"],
    "credit": ["Credit Account 1: $500", "Credit Account 2: $1,000"]
  };

  if (accounts[accountType]) {
    accountDetails = `Your ${accountType} accounts are: ${accounts[accountType].join(', ')}.`;
  } else {
    accountDetails = `Sorry, I couldn't find any details for ${accountType}.`;
  }


  const response = {
    sessionState: {
      dialogAction: {
        type: 'Close'
      },
      intent: {
        name: event.interpretations[0].intent.name,
        state: 'Fulfilled'
      }
    },
    messages: [
      {
        contentType: 'PlainText',
        content: accountDetails
      }
    ]
  };


  return response;
};
