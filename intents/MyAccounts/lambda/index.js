const AWS = require('aws-sdk');

exports.handler = async (event) => {
  const accountType = event.currentIntent.slots.AccountType.toLowerCase();
  let accountDetails = '';

  const accounts = {
    "checking account": ["Checking Account 1: $1,200", "Checking Account 2: $2,500"],
    "savings account": ["Savings Account 1: $5,000", "Savings Account 2: $10,000"],
    "credit account": ["Credit Account 1: $500", "Credit Account 2: $1,000"]
  };

  if (accounts[accountType]) {
    accountDetails = `Your ${accountType} accounts are: ${accounts[accountType].join(', ')}.`;
  } else {
    accountDetails = `Sorry, I couldn't find any details for ${accountType}.`;
  }

  const response = {
    dialogAction: {
      type: 'Close',
      fulfillmentState: 'Fulfilled',
      message: {
        contentType: 'PlainText',
        content: accountDetails
      }
    }
  };

  return response;
};
