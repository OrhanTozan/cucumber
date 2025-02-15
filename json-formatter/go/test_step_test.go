package json

import (
	"github.com/cucumber/messages-go/v15"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("ProcessTestStepFinished", func() {
	var (
		lookup *MessageLookup
	)

	BeforeEach(func() {
		lookup = &MessageLookup{}
		lookup.Initialize(false)

		pickleStep := &messages.PickleStep{
			Id:         "pickle-step-id",
			AstNodeIds: []string{"some-id"},
		}
		pickle := &messages.Pickle{
			Id:    "pickle-id",
			Steps: []*messages.PickleStep{pickleStep},
		}
		lookup.ProcessMessage(makePickleEnvelope(pickle))

		testStep := makeTestStep(
			"test-step-id",
			pickleStep.Id,
			[]string{},
		)

		testCase := makeTestCase(
			"test-case-id",
			pickle.Id,
			[]*messages.TestStep{testStep},
		)
		lookup.ProcessMessage(makeTestCaseEnvelope(testCase))

		testCaseStarted := &messages.TestCaseStarted{
			Id:         "test-case-started-id",
			TestCaseId: testCase.Id,
		}
		lookup.ProcessMessage(makeTestCaseStartedEnvelope(testCaseStarted))
	})

	It("has the TestCase ID", func() {
		testStepFinished := &messages.TestStepFinished{
			TestCaseStartedId: "test-case-started-id",
			TestStepId:        "test-step-id",
		}

		_, testStep := ProcessTestStepFinished(testStepFinished, lookup)
		Expect(testStep.TestCaseID).To(Equal("test-case-id"))
	})

	Context("with referencing issues", func() {
		It("returns nil if the step does not exist", func() {
			testStepFinished := &messages.TestStepFinished{
				TestCaseStartedId: "test-case-started-id",
				TestStepId:        "unknown-step",
			}
			_, testStep := ProcessTestStepFinished(testStepFinished, lookup)
			Expect(testStep).To(BeNil())
		})

		It("returns nil if the TestCaseStarted does not exist", func() {
			testStepFinished := &messages.TestStepFinished{
				TestCaseStartedId: "unknown-test-case-started",
				TestStepId:        "test-step-id",
			}
			_, testStep := ProcessTestStepFinished(testStepFinished, lookup)
			Expect(testStep).To(BeNil())
		})

		It("returns nil if the TestCaseStarted references an unknown TestCase", func() {
			testCaseStarted := &messages.TestCaseStarted{
				Id:         "test-case-started-no-test-case",
				TestCaseId: "unknown-test-case",
			}
			lookup.ProcessMessage(makeTestCaseStartedEnvelope(testCaseStarted))

			testStepFinished := &messages.TestStepFinished{
				TestCaseStartedId: testCaseStarted.Id,
				TestStepId:        "test-step-id",
			}
			_, testStep := ProcessTestStepFinished(testStepFinished, lookup)
			Expect(testStep).To(BeNil())
		})
	})

	Context("When step references a Hook", func() {
		BeforeEach(func() {
			hook := &messages.Hook{
				Id: "hook-id",
			}
			lookup.ProcessMessage(makeHookEnvelope(hook))

			testCase := makeTestCase(
				"test-case-id",
				"whatever-pickle-id",
				[]*messages.TestStep{
					makeHookTestStep("hook-step-id", hook.Id),
					makeHookTestStep("wrong-hook-step-id", "unknown-hook-id"),
				},
			)
			lookup.ProcessMessage(makeTestCaseEnvelope(testCase))

			testCaseStarted := &messages.TestCaseStarted{
				Id:         "hook-test-case-started-id",
				TestCaseId: testCase.Id,
			}
			lookup.ProcessMessage(makeTestCaseStartedEnvelope(testCaseStarted))
		})

		It("returns a TestStep including the Hook", func() {
			testStepFinished := &messages.TestStepFinished{
				TestCaseStartedId: "test-case-started-id",
				TestStepId:        "hook-step-id",
				TestStepResult: &messages.TestStepResult{
					Status: messages.TestStepResultStatus_PASSED,
				},
			}

			_, testStep := ProcessTestStepFinished(testStepFinished, lookup)

			Expect(testStep.Hook.Id).To(Equal("hook-id"))
			Expect(testStep.Result.Status).To(Equal(messages.TestStepResultStatus_PASSED))
		})

		It("returns a TestStep with a nil Step", func() {
			testStepFinished := &messages.TestStepFinished{
				TestCaseStartedId: "test-case-started-id",
				TestStepId:        "hook-step-id",
			}

			_, testStep := ProcessTestStepFinished(testStepFinished, lookup)

			Expect(testStep.Step).To(BeNil())
		})

		It("returns nil if the Hook does not exist", func() {
			testStepFinished := &messages.TestStepFinished{
				TestCaseStartedId: "test-case-started-id",
				TestStepId:        "wrong-hook-step-id",
			}
			_, testStep := ProcessTestStepFinished(testStepFinished, lookup)

			Expect(testStep).To(BeNil())
		})
	})

	Context("When step references a PickleStep", func() {
		var (
			testCaseStarted *messages.TestCaseStarted
			background      *messages.Background
		)
		BeforeEach(func() {
			// This is a bit dirty hack to avoid creating all the AST
			background = &messages.Background{
				Keyword: "Background",
			}
			backgroundStep := makeGherkinStep("background-step", "Given", "a passed step")
			step := makeGherkinStep("step-id", "Given", "a passed step")
			scenario := makeScenario("scenario-id", []*messages.Step{
				step,
			})
			lookup.stepByID[backgroundStep.Id] = backgroundStep
			lookup.stepByID[step.Id] = step
			lookup.scenarioByID[scenario.Id] = scenario
			lookup.backgroundByStepID[backgroundStep.Id] = background

			stepDefinitionConfig := &messages.StepDefinition{
				Id: "step-def-id",
				Pattern: &messages.StepDefinitionPattern{
					Source: "a passed {word}",
				},
			}
			lookup.ProcessMessage(makeStepDefinitionEnvelope(stepDefinitionConfig))

			backgroundPickleStep := &messages.PickleStep{
				Id:         "background-pickle-step-id",
				AstNodeIds: []string{backgroundStep.Id},
				Text:       "a passed step",
			}

			pickleStep := &messages.PickleStep{
				Id:         "pickle-step-id",
				AstNodeIds: []string{step.Id},
				Text:       "a passed step",
			}

			pickle := &messages.Pickle{
				Id:         "pickle-id",
				Uri:        "some_feature.feature",
				AstNodeIds: []string{scenario.Id},
				Steps: []*messages.PickleStep{
					backgroundPickleStep,
					pickleStep,
				},
			}
			lookup.ProcessMessage(makePickleEnvelope(pickle))

			testCase := makeTestCase(
				"test-case-id",
				pickle.Id,
				[]*messages.TestStep{
					makeTestStep("background-step-id", backgroundPickleStep.Id, []string{"step-def-id"}),
					makeTestStep("test-step-id", pickleStep.Id, []string{"step-def-id"}),
					makeTestStep("unknown-pickle", "unknown-pickle-step-id", []string{}),
				},
			)
			lookup.ProcessMessage(makeTestCaseEnvelope(testCase))

			testCaseStarted = &messages.TestCaseStarted{
				Id:         "test-case-started-id",
				TestCaseId: testCase.Id,
			}
			lookup.ProcessMessage(makeTestCaseStartedEnvelope(testCaseStarted))
		})

		It("returns a TestStep including the FeatureStep", func() {
			testStepFinished := &messages.TestStepFinished{
				TestStepId:        "test-step-id",
				TestCaseStartedId: testCaseStarted.Id,
			}

			_, testStep := ProcessTestStepFinished(testStepFinished, lookup)
			Expect(testStep.Step.Id).To(Equal("step-id"))
		})

		It("returns a Step including the StepDefinitions", func() {
			testStepFinished := &messages.TestStepFinished{
				TestStepId:        "test-step-id",
				TestCaseStartedId: testCaseStarted.Id,
			}
			_, testStep := ProcessTestStepFinished(testStepFinished, lookup)
			Expect(len(testStep.StepDefinitions)).To(Equal(1))
			Expect(testStep.StepDefinitions[0].Pattern.Source).To(Equal("a passed {word}"))
		})

		It("the Step is not marked as a Background step by default", func() {
			testStepFinished := &messages.TestStepFinished{
				TestStepId:        "test-step-id",
				TestCaseStartedId: testCaseStarted.Id,
			}
			_, testStep := ProcessTestStepFinished(testStepFinished, lookup)
			Expect(testStep.Background).To(BeNil())
		})

		Context("when the Step is defined in a background", func() {
			It("sets Step.IsBackground to the Background message", func() {
				testStepFinished := &messages.TestStepFinished{
					TestStepId:        "background-step-id",
					TestCaseStartedId: testCaseStarted.Id,
				}
				_, testStep := ProcessTestStepFinished(testStepFinished, lookup)
				Expect(testStep.Background).To(Equal(background))
			})
		})

		Context("with referencing issues", func() {
			It("returns Nil if the Pickle is unknow", func() {
				testCase := makeTestCase(
					"test-case-id",
					"unknown-pickle",
					[]*messages.TestStep{},
				)
				lookup.ProcessMessage(makeTestCaseEnvelope(testCase))

				testCaseStarted := &messages.TestCaseStarted{
					Id:         "test-case-started-id",
					TestCaseId: testCase.Id,
				}
				lookup.ProcessMessage(makeTestCaseStartedEnvelope(testCaseStarted))

				testStepFinished := &messages.TestStepFinished{
					TestStepId:        "test-step-id",
					TestCaseStartedId: testCaseStarted.Id,
				}

				_, testStep := ProcessTestStepFinished(testStepFinished, lookup)
				Expect(testStep).To(BeNil())
			})

			It("Returns Nil if the PickleStep is unknown", func() {
				testStepFinished := &messages.TestStepFinished{
					TestStepId:        "unknown-pickle-step",
					TestCaseStartedId: testCaseStarted.Id,
				}

				_, testStep := ProcessTestStepFinished(testStepFinished, lookup)
				Expect(testStep).To(BeNil())
			})
		})
	})
})

var _ = Describe("TestStepToJSON", func() {
	var (
		step     *TestStep
		jsonStep *jsonStep
	)

	Context("When TestStep comes from a Hook", func() {
		BeforeEach(func() {
			step = &TestStep{
				Hook: &messages.Hook{
					SourceReference: &messages.SourceReference{
						Uri: "some/hooks.go",
						Location: &messages.Location{
							Column: 3,
							Line:   12,
						},
					},
				},
				Result: &messages.TestStepResult{
					Status: messages.TestStepResultStatus_PASSED,
					Duration: &messages.Duration{
						Seconds: 123,
						Nanos:   456,
					},
				},
				Attachments: make([]*messages.Attachment, 0),
			}

			step.Attachments = append(step.Attachments, &messages.Attachment{
				Body:            "Hello",
				MediaType:       "text/plain",
				ContentEncoding: messages.AttachmentContentEncoding_BASE64,
			})

			jsonStep = TestStepToJSON(step)
		})

		It("Has a Match", func() {
			Expect(jsonStep.Match.Location).To(Equal("some/hooks.go:12"))
		})

		It("Has a Result", func() {
			Expect(jsonStep.Result.Status).To(Equal("passed"))
		})

		It("Has a Duration", func() {
			Expect(jsonStep.Result.Duration).To(Equal(uint64(123000000456)))
		})

		It("has an Embedding", func() {
			Expect(len(jsonStep.Embeddings)).To(Equal(1))
			Expect(jsonStep.Embeddings[0].MimeType).To(Equal("text/plain"))
			Expect(jsonStep.Embeddings[0].Data).To(Equal("Hello"))
		})
	})

	Context("When TestStep comes from a feature step", func() {
		BeforeEach(func() {
			step = &TestStep{
				Step: &messages.Step{
					Id:      "some-id",
					Keyword: "Given",
					Text:    "a <status> step",
					Location: &messages.Location{
						Line: 5,
					},
				},
				Pickle: &messages.Pickle{
					Uri: "my_feature.feature",
				},
				PickleStep: &messages.PickleStep{
					Text: "a passed step",
				},
				Result: &messages.TestStepResult{
					Status: messages.TestStepResultStatus_FAILED,
					Duration: &messages.Duration{
						Seconds: 123,
						Nanos:   456,
					},
				},
				Attachments: make([]*messages.Attachment, 0),
			}

			step.Attachments = append(step.Attachments, &messages.Attachment{
				Body:            "Hello",
				MediaType:       "text/plain",
				ContentEncoding: messages.AttachmentContentEncoding_BASE64,
			})

			jsonStep = TestStepToJSON(step)
		})

		It("gets keyword from Step", func() {
			Expect(jsonStep.Keyword).To(Equal("Given"))
		})

		It("should gets name from PickleStep", func() {
			Expect(jsonStep.Name).To(Equal("a passed step"))
		})

		It("has a Result", func() {
			Expect(jsonStep.Result.Status).To(Equal("failed"))
		})

		It("Has a Duration", func() {
			Expect(jsonStep.Result.Duration).To(Equal(uint64(123000000456)))
		})

		It("has a Line", func() {
			Expect(jsonStep.Line).To(Equal(uint32(5)))
		})

		It("has an Embedding", func() {
			Expect(len(jsonStep.Embeddings)).To(Equal(1))
			Expect(jsonStep.Embeddings[0].MimeType).To(Equal("text/plain"))
			Expect(jsonStep.Embeddings[0].Data).To(Equal("Hello"))
		})

		Context("When it does not have a StepDefinition", func() {
			It("Has a Match referencing the feature file", func() {
				Expect(jsonStep.Match.Location).To(Equal("my_feature.feature:5"))
			})
		})

		Context("When it has a StepDefinition", func() {
			It("has a Match referencing the feature file", func() {
				step.StepDefinitions = []*messages.StepDefinition{
					&messages.StepDefinition{
						SourceReference: &messages.SourceReference{
							Uri: "support_code.go",
							Location: &messages.Location{
								Line: 12,
							},
						},
					},
				}

				jsonStep = TestStepToJSON(step)
				Expect(jsonStep.Match.Location).To(Equal("support_code.go:12"))
			})
		})
	})

	Context("When SourceReference uses JavaMethod and no Location", func() {
		Describe("from a Hook", func() {
			BeforeEach(func() {
				step = &TestStep{
					Hook: &messages.Hook{
						SourceReference: &messages.SourceReference{
							JavaMethod: &messages.JavaMethod{
								ClassName:  "org.cucumber.jvm.Class",
								MethodName: "someMethod",
								MethodParameterTypes: []string{
									"java.lang.String",
									"int",
								},
							},
						},
					},
					Result: &messages.TestStepResult{
						Status: messages.TestStepResultStatus_PASSED,
						Duration: &messages.Duration{
							Seconds: 123,
							Nanos:   456,
						},
					},
				}
				jsonStep = TestStepToJSON(step)
			})

			It("Has a Match", func() {
				Expect(jsonStep.Match.Location).To(Equal("org.cucumber.jvm.Class.someMethod(java.lang.String,int)"))
			})
		})

		Describe("from a feature step", func() {
			BeforeEach(func() {
				step = &TestStep{
					Step: &messages.Step{
						Id:      "some-id",
						Keyword: "Given",
						Text:    "a <status> step",
						Location: &messages.Location{
							Line: 5,
						},
					},
					Pickle: &messages.Pickle{
						Uri: "my_feature.feature",
					},
					PickleStep: &messages.PickleStep{
						Text: "a passed step",
					},
					Result: &messages.TestStepResult{
						Status: messages.TestStepResultStatus_FAILED,
						Duration: &messages.Duration{
							Seconds: 123,
							Nanos:   456,
						},
					},
					StepDefinitions: []*messages.StepDefinition{
						&messages.StepDefinition{
							SourceReference: &messages.SourceReference{
								JavaMethod: &messages.JavaMethod{
									ClassName:  "org.cucumber.jvm.Class",
									MethodName: "someMethod",
									MethodParameterTypes: []string{
										"java.lang.String",
										"int",
									},
								},
							},
						},
					},
				}
				jsonStep = TestStepToJSON(step)
			})

			It("Has a Match", func() {
				Expect(jsonStep.Match.Location).To(Equal("org.cucumber.jvm.Class.someMethod(java.lang.String,int)"))
			})
		})
	})

	Context("When SourceReference uses JavaStackTraceElement and Location", func() {
		Describe("from a Hook", func() {
			BeforeEach(func() {
				step = &TestStep{
					Hook: &messages.Hook{
						SourceReference: &messages.SourceReference{
							JavaStackTraceElement: &messages.JavaStackTraceElement{
								ClassName:  "org.cucumber.jvm.ExceptionClass",
								MethodName: "someMethod",
								FileName:   "ExceptionClass.java",
							},
							Location: &messages.Location{
								Line: 123,
							},
						},
					},
					Result: &messages.TestStepResult{
						Status: messages.TestStepResultStatus_PASSED,
						Duration: &messages.Duration{
							Seconds: 123,
							Nanos:   456,
						},
					},
				}
				jsonStep = TestStepToJSON(step)
			})

			It("Has a Match", func() {
				Expect(jsonStep.Match.Location).To(Equal("ExceptionClass.java:123"))
			})
		})

		Describe("from a feature step", func() {
			BeforeEach(func() {
				step = &TestStep{
					Step: &messages.Step{
						Id:      "some-id",
						Keyword: "Given",
						Text:    "a <status> step",
						Location: &messages.Location{
							Line: 5,
						},
					},
					Pickle: &messages.Pickle{
						Uri: "my_feature.feature",
					},
					PickleStep: &messages.PickleStep{
						Text: "a passed step",
					},
					Result: &messages.TestStepResult{
						Status: messages.TestStepResultStatus_FAILED,
						Duration: &messages.Duration{
							Seconds: 123,
							Nanos:   456,
						},
					},
					StepDefinitions: []*messages.StepDefinition{
						&messages.StepDefinition{
							SourceReference: &messages.SourceReference{
								JavaStackTraceElement: &messages.JavaStackTraceElement{
									ClassName:  "org.cucumber.jvm.ExceptionClass",
									MethodName: "someMethod",
									FileName:   "ExceptionClass.java",
								},
								Location: &messages.Location{
									Line: 123,
								},
							},
						},
					},
				}
				jsonStep = TestStepToJSON(step)
			})

			It("Has a Match", func() {
				Expect(jsonStep.Match.Location).To(Equal("ExceptionClass.java:123"))
			})
		})
	})

	Context("When SourceReference uses JavaStackTraceElement and no Location", func() {
		Describe("from a Hook", func() {
			BeforeEach(func() {
				step = &TestStep{
					Hook: &messages.Hook{
						SourceReference: &messages.SourceReference{
							JavaStackTraceElement: &messages.JavaStackTraceElement{
								ClassName:  "org.cucumber.jvm.ExceptionClass",
								MethodName: "someMethod",
								FileName:   "ExceptionClass.java",
							},
						},
					},
					Result: &messages.TestStepResult{
						Status: messages.TestStepResultStatus_PASSED,
						Duration: &messages.Duration{
							Seconds: 123,
							Nanos:   456,
						},
					},
				}
				jsonStep = TestStepToJSON(step)
			})

			It("Has a Match", func() {
				Expect(jsonStep.Match.Location).To(Equal("ExceptionClass.java"))
			})
		})

		Describe("from a feature step", func() {
			BeforeEach(func() {
				step = &TestStep{
					Step: &messages.Step{
						Id:      "some-id",
						Keyword: "Given",
						Text:    "a <status> step",
						Location: &messages.Location{
							Line: 5,
						},
					},
					Pickle: &messages.Pickle{
						Uri: "my_feature.feature",
					},
					PickleStep: &messages.PickleStep{
						Text: "a passed step",
					},
					Result: &messages.TestStepResult{
						Status: messages.TestStepResultStatus_FAILED,
						Duration: &messages.Duration{
							Seconds: 123,
							Nanos:   456,
						},
					},
					StepDefinitions: []*messages.StepDefinition{
						&messages.StepDefinition{
							SourceReference: &messages.SourceReference{
								JavaStackTraceElement: &messages.JavaStackTraceElement{
									ClassName:  "org.cucumber.jvm.ExceptionClass",
									MethodName: "someMethod",
									FileName:   "ExceptionClass.java",
								},
							},
						},
					},
				}
				jsonStep = TestStepToJSON(step)
			})

			It("Has a Match", func() {
				Expect(jsonStep.Match.Location).To(Equal("ExceptionClass.java"))
			})
		})
	})

})
